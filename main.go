package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shipyard-run/shipyard/pkg/clients"
	flag "github.com/spf13/pflag"

	"github.com/hashicorp/go-hclog"
)

var ports = flag.StringSlice("ports", nil, "Specify the ports to forward [Local]:[Remote] 8081:8080")
var serviceName = flag.String("service-name", "", "FQDN of the service in the Docker network, Nomad job.group.task, or Kubernetes pod/service")
var proxyType = flag.String("proxy-type", "docker", "Type of proxy to use [docker | kubernetes | nomad]")
var namespace = flag.String("namespace", "default", "Namespace for Kubernetes services")
var nomadConfig = flag.String("nomad-config", "~/.shipyard/nomad.json", "Config filefor Nomad server")

func main() {
	flag.Parse()

	log := hclog.Default()

	log.Info("Starting Ingress")

	p, err := splitPorts(*ports)
	if err != nil {
		log.Error("Error parsing parameters", "error", err)
		os.Exit(1)
	}

	for {
		switch *proxyType {
		case "docker":
			err = createNetCat(*serviceName, p, log)
		case "kubernetes":
			err = createKubeProxy(*serviceName, p, log)
		case "nomad":
			err = createNomadProxy(*serviceName, *nomadConfig, p, log)
		}

		if err != nil {
			log.Error("Error creating connection, retrying", "error", err)
			time.Sleep(2 * time.Second)
		}
	}
}

func createNetCat(service string, ports [][]string, log hclog.Logger) error {
	errChan := make(chan error)

	for _, p := range ports {
		log.Info("Running socat", "address", service, "local", p[0], "remote", p[1])

		c := exec.Command(
			"socat",
			fmt.Sprintf("tcp-l:%s,fork,reuseaddr", p[0]),
			fmt.Sprintf("tcp:%s:%s", service, p[1]),
		)

		// set the standard out and error to the logger
		c.Stdout = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
		c.Stderr = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})

		go func() {
			errChan <- c.Run()
		}()
	}

	return <-errChan
}

func createKubeProxy(service string, ports [][]string, log hclog.Logger) error {
	if len(ports) < 1 {
		return fmt.Errorf("Please specify at least 1 port mapping for Kubernetes proxies")
	}

	args := []string{
		"port-forward",
		"-n",
		*namespace,
		service,
		"--address",
		"0.0.0.0",
	}

	for _, p := range ports {
		args = append(args, fmt.Sprintf("%s:%s", p[0], p[1]))
	}

	c := exec.Command(
		"kubectl",
		args...,
	)

	// set the standard out and error to the logger
	errorChan := make(chan error)
	doneChan := make(chan error)
	c.Stdout = &HijackWriter{log, nil}
	c.Stderr = &HijackWriter{log.Named("error_log"), errorChan}

	go func() {
		doneChan <- c.Run()
	}()

	select {
	case err := <-doneChan:
		return err
	case err := <-errorChan:
		// kill the run process and exit
		c.Process.Kill()
		return err
	}
}

func createNomadProxy(service string, configlocation string, ports [][]string, log hclog.Logger) error {
	if len(ports) != 1 {
		return fmt.Errorf("Please specify only 1 port mapping for Nomad proxies")
	}

	// check the service, should be job.group.task
	parts := strings.Split(service, ".")
	if len(parts) != 3 {
		return fmt.Errorf("Service should be specified as job.group.task, got: %s", service)
	}

	// lookup the endpoints
	log.Info("Querying endpoints in Nomad server using", "config", configlocation)
	http := clients.NewHTTP(2*time.Second, log)
	client := clients.NewNomad(http, 2*time.Second, log)

	err := client.SetConfig(configlocation, string(clients.RemoteContext))
	if err != nil {
		return fmt.Errorf("Unable to set nomad config: %s", err)
	}

	ep, err := client.Endpoints(parts[0], parts[1], parts[2])
	if err != nil {
		return fmt.Errorf("Unable to check endpoints: %s", err)
	}

	if len(ep) == 0 {
		return fmt.Errorf("No endpoints found for service: %s", service)
	}

	log.Info("Got endpoints", "endpoints", ep)

	// we have endpoints find a matching port
	uris := []string{}

	// check there is an endpoint for the given port
	for _, e := range ep {
		if v, ok := e[ports[0][1]]; ok {
			uris = append(uris, v)
		}
	}

	if len(uris) == 0 {
		return fmt.Errorf("No endpoints found for service: %s, and port: %s", service, ports[0][1])
	}

	randEP := rand.Intn(len(uris))
	portsPair := strings.Split(uris[randEP], ":")

	// Create the port list using the endpoint port and the mapped local
	ports = [][]string{
		[]string{
			ports[0][0],
			portsPair[1],
		},
	}

	// start socat
	return createNetCat(portsPair[0], ports, log)
}

// HijackWriter is a simple writer which brodcasts to a channel when a log message is called
type HijackWriter struct {
	log        hclog.Logger
	notifyChan chan error
}

func (h *HijackWriter) Write(p []byte) (n int, err error) {
	h.log.Info(string(p))

	// notify that we have logged a message
	if h.notifyChan != nil {
		h.notifyChan <- fmt.Errorf("%s", p)
	}

	return len(p), nil
}

func splitPorts(ports []string) ([][]string, error) {
	rp := [][]string{}

	for _, p := range ports {
		pp := strings.Split(p, ":")
		if len(pp) != 2 {
			return nil, fmt.Errorf("Error ports should be specified as a : separated string [local]:[remote]")
		}

		rp = append(rp, pp)
	}

	return rp, nil
}
