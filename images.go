package docker

import (
	"fmt"

	"github.com/farseer-go/utils/exec"
)

type images struct {
	//progress chan string
}

// Pull 拉取镜像
func (receiver images) Pull(image string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker pull %s", image), nil, "", true)
}

// ClearImages 清除镜像
func (receiver images) ClearImages() (chan string, func() int) {
	// docker image prune -a -f 这个是安全的删除
	return exec.RunShell(`docker system prune -a -f`, nil, "", false)
}
