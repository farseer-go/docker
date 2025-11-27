package docker

import (
	"fmt"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
)

type images struct {
	progress chan string
}

// Pull 拉取镜像
func (receiver images) Pull(image string) error {
	exitCode := exec.RunShell(fmt.Sprintf("docker pull %s", image), receiver.progress, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// ClearImages 清除镜像
func (receiver images) ClearImages() error {
	// docker image prune -a -f 这个是安全的删除
	var exitCode = exec.RunShell(`docker system prune -a -f`, receiver.progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}
