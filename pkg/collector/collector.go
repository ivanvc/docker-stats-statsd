package collector

import (
	"github.com/ivanvc/docker-stats-statsd/pkg/docker"

	"github.com/statsd/client-interface"
	"github.com/statsd/client-namespace"
)

type Collector struct {
	Containers []*docker.Container
	client     statsd.Client
}

func New(cl statsd.Client) *Collector {
	return &Collector{client: cl}
}

func (c *Collector) Start() {
	for _, con := range c.Containers {
		con.Start(namespace.New(c.client, con.Name))
	}
}

func (c *Collector) Stop() {
	for _, con := range c.Containers {
		con.Stop()
	}
	c.client.Flush()
}
