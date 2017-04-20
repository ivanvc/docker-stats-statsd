package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ivanvc/docker-stats-statsd/pkg/collector"
	"github.com/ivanvc/docker-stats-statsd/pkg/docker"

	"github.com/statsd/client"
	"github.com/statsd/client-namespace"
)

var (
	dockerURI    = os.Getenv("DOCKER_API_URI")
	statsdHost   = os.Getenv("STATSD_HOST") + ":" + os.Getenv("STATSD_PORT")
	statsdPrefix = os.Getenv("STATSD_PREFIX")
)

func main() {
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	client, err := statsd.Dial(statsdHost)
	if err != nil {
		panic(err)
	}

	docker.C = docker.NewClient(dockerURI)
	c := collector.New(namespace.New(client, statsdPrefix))
	c.Start()

	<-sigTerm
	c.Stop()
}
