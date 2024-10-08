package docker

import (
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
)

type images struct {
}

// Pull 拉取镜像
func (receiver images) Pull(image string) error {
	c := make(chan string, 100)
	exitCode := exec.RunShell(fmt.Sprintf("docker pull %s", image), c, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// ClearImages 清除镜像
func (receiver images) ClearImages() error {
	c := make(chan string, 100)
	var exitCode = exec.RunShell(`docker system prune -a -f`, c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}
