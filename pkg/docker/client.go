package docker

import "github.com/fsouza/go-dockerclient"

var C *docker.Client

func NewClient(f string) *docker.Client {
	client, err := docker.NewClient(f)
	if err != nil {
		panic(err)
	}
	return client
}
