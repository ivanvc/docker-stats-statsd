package container

import (
	"log"
	"sync"
	"time"

	d "github.com/ivanvc/docker-stats-statsd/pkg/docker"

	"github.com/fsouza/go-dockerclient"
	"github.com/statsd/client-interface"
)

const (
	DiscoveryTimeout = 5 * time.Minute
	ReportTimeout    = time.Second
)

type Container struct {
	sync.Mutex

	Name string
	ID   string

	quitChan     chan interface{}
	Stalled      bool
	timeout      <-chan time.Time
	reportTicker *time.Ticker
	stats        *docker.Stats

	doneChan  chan bool
	statsChan chan *docker.Stats
	statsdNs  statsd.Client
}

func List() []*Container {
	containers, err := d.C.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	cons := make([]*Container, len(containers))
	for i, c := range containers {
		cons[i] = &Container{
			ID:   c.ID,
			Name: c.Names[0][1:],
		}
	}
	return cons
}

func (c *Container) Start(ns statsd.Client) {
	c.quitChan = make(chan interface{})
	c.doneChan = make(chan bool)
	c.statsChan = make(chan *docker.Stats)
	c.statsdNs = ns
	c.reportTicker = time.NewTicker(ReportTimeout)

	log.Printf("Container %s(%s) registered", c.Name, c.ID)

	go d.C.Stats(docker.StatsOptions{
		Done:   c.doneChan,
		ID:     c.ID,
		Stats:  c.statsChan,
		Stream: true,
	})

	go c.report()
}

func (c *Container) report() {
	c.KeepAlive()
	for {
		select {
		case <-c.timeout:
			c.Stop()
		case s := <-c.statsChan:
			if s == nil {
				c.Stop()
			} else {
				c.Lock()
				c.stats = s
				c.Unlock()
			}
		case <-c.reportTicker.C:
			c.Lock()
			cpuDelta := float64(c.stats.CPUStats.CPUUsage.TotalUsage) - float64(c.stats.PreCPUStats.CPUUsage.TotalUsage)
			sysDelta := float64(c.stats.CPUStats.SystemCPUUsage) - float64(c.stats.PreCPUStats.SystemCPUUsage)
			cpuPercent := (cpuDelta / sysDelta) * float64(len(c.stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
			memPercent := float64(c.stats.MemoryStats.Usage) / float64(c.stats.MemoryStats.Limit) * 100.0
			c.statsdNs.Gauge("memory.percent", int(memPercent))
			c.statsdNs.Gauge("cpu.percent", int(cpuPercent))
			c.Unlock()
		case <-c.quitChan:
			c.doneChan <- true
			return
		}
	}
}

func (c *Container) Stop() {
	log.Printf("Stopped collection of %s(%s) container", c.Name, c.ID)

	c.Stalled = true
	c.reportTicker.Stop()
	close(c.quitChan)
}

func (c *Container) KeepAlive() {
	c.Stalled = false
	c.timeout = time.After(2 * DiscoveryTimeout)
}
