package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/farseer-go/fs/flog"
)

// Client docker client
type Client struct {
	dockerClient *client.Client
	Container    container
	Service      service
	Node         node
	Hub          hub
	Images       images
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
		Node:         node{dockerClient: cli},
		Hub:          hub{dockerClient: cli},
		Images:       images{dockerClient: cli},
	}, nil
}

// GetVersion 获取系统Docker版本
func (receiver Client) GetVersion() string {
	version, err := receiver.dockerClient.ServerVersion(context.Background())
	if err != nil {
		flog.Warning(err.Error())
		return ""
	}
	return version.Version
}
