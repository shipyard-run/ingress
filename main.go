package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/hashicorp/go-hclog"
)

var ports = flag.StringSlice("ports", nil, "Specify the ports to forward [Local]:[Remote] 8081:8080")
var serviceName = flag.String("service-name", "", "FQDN of the service in the Docker network or Kubernetes pod/service")
var proxyType = flag.String("proxy-type", "docker", "Type of proxy to use [docker | kubernetes]")

func main() {
	flag.Parse()

	log := hclog.Default()

	log.Info("Starting Ingress")

	p, err := splitPorts(*ports)
	if err != nil {
		log.Error("Error parsing parameters", "error", err)
		os.Exit(1)
	}

	if *proxyType == "docker" {
		err = createNetCat(*serviceName, p, log)
	} else {
		err = createKubeProxy(*serviceName, p, log)
	}

	if err != nil {
		log.Error("Error creating connection", "error", err)
		os.Exit(1)
	}
}

func createNetCat(service string, ports [][]string, log hclog.Logger) error {
	if len(ports) != 1 {
		return fmt.Errorf("Please specify a single port mapping for Docker proxies")
	}

	c := exec.Command(
		"socat",
		fmt.Sprintf("tcp-l:%s,fork,reuseaddr", ports[0][0]),
		fmt.Sprintf("tcp:%s:%s", service, ports[0][1]),
	)

	// set the standard out and error to the logger
	c.Stdout = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
	c.Stderr = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})

	return c.Run()
}

func createKubeProxy(service string, ports [][]string, log hclog.Logger) error {
	if len(ports) < 1 {
		return fmt.Errorf("Please specify at least 1 port mapping for Kubernetes proxies")
	}

	args := []string{
		"port-forward",
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
	c.Stdout = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
	c.Stderr = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})

	return c.Run()
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
