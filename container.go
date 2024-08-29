package docker

import (
	"context"
	"github.com/docker/docker/client"
)

type container struct {
	dockerClient *client.Client
}

// Exists 判断容器是否已创建
func (receiver container) Exists(containerId string) bool {
	inspect, err := receiver.dockerClient.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return false
	}
	return inspect.Name == "/"+containerId
	/*
		progress := make(chan string, 1000)
		// docker inspect fops
		var exitCode = exec.RunShell(fmt.Sprintf("docker inspect %s", dockerName), progress, nil, "", false)
		lst := collections.NewListFromChan(progress)
		if exitCode != 0 {
			if lst.Contains("[]") && lst.ContainsPrefix("Error: No such object:") {
				return false
			}
			return false
		}
		if lst.Contains("[]") && lst.ContainsPrefix("Error: No such object:") {
			return false
		}
		return lst.ContainsAny(fmt.Sprintf("\"Name\": \"/%s\",", dockerName))
	*/
}
