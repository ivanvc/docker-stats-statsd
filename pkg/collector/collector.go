package collector

import (
	"log"
	"time"

	"github.com/ivanvc/docker-stats-statsd/pkg/container"

	"github.com/statsd/client-interface"
	"github.com/statsd/client-namespace"
)

const DiscoveryTimeout = 5 * time.Minute

type Collector struct {
	containers      map[string]*container.Container
	client          statsd.Client
	quitChan        chan interface{}
	discoveryTicker *time.Ticker
}

func New(cl statsd.Client) *Collector {
	return &Collector{
		client:     cl,
		containers: make(map[string]*container.Container),
	}
}

func (c *Collector) Start() {
	c.quitChan = make(chan interface{})
	c.discoverContainers()

	c.discoveryTicker = time.NewTicker(DiscoveryTimeout)
	go c.collect()
}

func (c *Collector) collect() {
	for {
		select {
		case <-c.discoveryTicker.C:
			c.discoverContainers()
			c.removeStalledContainers()
		case <-c.quitChan:
			return
		}
	}
}

func (c *Collector) discoverContainers() {
	log.Println("Discoverying new containers")
	for _, con := range container.List() {
		if r, ok := c.containers[con.ID]; ok {
			r.KeepAlive()
			continue
		}
		c.containers[con.ID] = con
		con.Start(namespace.New(c.client, con.Name))
	}
}

func (c *Collector) removeStalledContainers() {
	for id, con := range c.containers {
		if con.Stalled {
			delete(c.containers, id)
		}
	}
}

func (c *Collector) Stop() {
	log.Println("Stopping collector")
	close(c.quitChan)
	c.discoveryTicker.Stop()
	for _, con := range c.containers {
		con.Stop()
	}
	c.client.Flush()
}
