package stats

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	sockFile string
}

func NewDocker(f string) *Docker {
	return &Docker{f}
}

func (d *Docker) Start() {
	client, err := docker.NewClient(d.sockFile)
	if err != nil {
		panic(err)
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		s := make(chan *docker.Stats)
		go func(id string) {
			go client.Stats(docker.StatsOptions{
				ID:     id,
				Stats:  s,
				Stream: true,
			})
			for {
				st := <-s
				cpuDelta := float64(st.CPUStats.CPUUsage.TotalUsage) - float64(st.PreCPUStats.CPUUsage.TotalUsage)
				sysDelta := float64(st.CPUStats.SystemCPUUsage) - float64(st.PreCPUStats.SystemCPUUsage)
				cpuPercent := (cpuDelta / sysDelta) * float64(len(st.CPUStats.CPUUsage.PercpuUsage)) * 100.0
				memPercent := float64(st.MemoryStats.Usage) / float64(st.MemoryStats.Limit) * 100.0
				fmt.Printf("%s: Mem: %.2f%% CPU: %.2f%%\n", id[:7], memPercent, cpuPercent)
			}
		}(c.ID)
	}
	for {
	}
}
