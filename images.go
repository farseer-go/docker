package docker

import (
	"fmt"
	"net/http"

	"github.com/farseer-go/utils/exec"
)

type images struct {
	unixClient *http.Client
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
