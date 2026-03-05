package docker

import (
	"net/http"
	"strings"

	"fmt"
)

type task struct {
	unixClient *http.Client
}

// Inspect 查看服务详情
// 注意：由于 Service 本身不包含 ContainerID，此方法实际上是通过 Service ID 查询其关联的 Task 列表
func (receiver task) Inspect(taskId string) (ServiceIdInspectJson, error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/tasks/kbo9xu9qtxw1b69r02tt57bvh
	url := fmt.Sprintf("http://localhost/tasks/%s", taskId)

	// 调用工具方法解析
	task, err := UnixGetDecode[ServiceIdInspectJson](receiver.unixClient, url)
	if err != nil {
		return task, err
	}

	// 截断 ContainerID，保持和 CLI 输出一致
	if len(task.Status.ContainerStatus.ContainerID) >= 12 {
		task.Status.ContainerStatus.ContainerID = task.Status.ContainerStatus.ContainerID[:12]
	}

	task.Spec.ContainerSpec.Image = strings.Split(task.Spec.ContainerSpec.Image, "@")[0] // 去掉 digest 部分

	return task, nil
}
