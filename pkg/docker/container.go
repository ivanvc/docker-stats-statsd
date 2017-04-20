package docker

import (
	dc "github.com/fsouza/go-dockerclient"
	"github.com/statsd/client-interface"
)

type Container struct {
	Name     string
	id       string
	client   *dc.Client
	quitChan chan bool
}

func GetContainers(f string) []*Container {
	client, err := dc.NewClient(f)
	if err != nil {
		panic(err)
	}
	containers, err := client.ListContainers(dc.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	cons := make([]*Container, len(containers))
	for i, c := range containers {
		cons[i] = &Container{
			id:     c.ID,
			Name:   c.Names[0][1:],
			client: client,
		}
	}
	return cons
}

func (c *Container) Start(ns statsd.Client) {
	d := make(chan bool)
	s := make(chan *dc.Stats)
	go c.client.Stats(dc.StatsOptions{
		ID:     c.id,
		Stats:  s,
		Stream: true,
		Done:   d,
	})

	go func() {
		for {
			select {
			case st := <-s:
				cpuDelta := float64(st.CPUStats.CPUUsage.TotalUsage) - float64(st.PreCPUStats.CPUUsage.TotalUsage)
				sysDelta := float64(st.CPUStats.SystemCPUUsage) - float64(st.PreCPUStats.SystemCPUUsage)
				cpuPercent := (cpuDelta / sysDelta) * float64(len(st.CPUStats.CPUUsage.PercpuUsage)) * 100.0
				memPercent := float64(st.MemoryStats.Usage) / float64(st.MemoryStats.Limit) * 100.0
				ns.Gauge("memory.percent", int(memPercent))
				ns.Gauge("cpu.percent", int(cpuPercent))
			case <-c.quitChan:
				d <- true
				return
			}
		}
	}()
}

func (c *Container) Stop() {
	go func() {
		c.quitChan <- true
	}()
}
