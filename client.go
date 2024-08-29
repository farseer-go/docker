package docker

import (
	"github.com/docker/docker/client"
)

// Client docker client
type Client struct {
	dockerClient *client.Client
	Container    container
	Service      service
}

// NewClient 实例化一个Client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{
		dockerClient: cli,
		Container:    container{dockerClient: cli},
		Service:      service{dockerClient: cli},
	}, nil
}
