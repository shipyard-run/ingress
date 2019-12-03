package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
)

var portHost = flag.Int("port-host", -1, "Specify the port to expose on the host machine")
var portRemote = flag.Int("port-remote", -1, "Specify the port where the service is listening on the remote service")
var serviceName = flag.String("service-name", "", "FQDN of the service in the docker or k8s network")

func main() {
	flag.Parse()

	log := hclog.Default()

	log.Info("Starting Ingress")

	err := createNetCat(*serviceName, *portHost, *portRemote, log)
	if err != nil {
		log.Error("Error creating connection", "error", err)
		os.Exit(1)
	}
}

func createNetCat(service string, portRemote, portHost int, log hclog.Logger) error {
	c := exec.Command(
		"socat",
		fmt.Sprintf("tcp-l:%d,fork,reuseaddr", portHost),
		fmt.Sprintf("tcp:%s:%d", service, portRemote),
	)

	// set the standard out and error to the logger
	c.Stdout = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})
	c.Stderr = log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true})

	return c.Run()
}
