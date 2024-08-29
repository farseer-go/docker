package docker

import "github.com/docker/docker/client"

type service struct {
	dockerClient *client.Client
}
