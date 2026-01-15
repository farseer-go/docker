package docker

import (
	"github.com/farseer-go/fs/async"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type event struct {
	//progress chan string
}

// Watch 持续获取docker事件
func (receiver event) Watch() chan EventResult {
	eventResultChan := make(chan EventResult, 1000)
	progress, wait := exec.RunShell("docker events --format '{{json .}}'", nil, "", false)

	// 将读取到的json事件信息转换成EventResult结构体
	worker := async.New()
	worker.Add(func() {
		for json := range progress {
			var eventResult EventResult
			snc.Unmarshal([]byte(json), &eventResult)
			eventResultChan <- eventResult
		}
	})
	defer worker.Wait()

	wait()

	return eventResultChan
}

// docker 事件结构体
type EventResult struct {
	Status string `json:"status"` // 事件类型
	ID     string `json:"id"`     // 容器ID
	From   string `json:"from"`   // 镜像
	Type   string `json:"Type"`   // container
	Action string `json:"Action"` // start
	Actor  struct {
		ID         string `json:"ID"` // 容器ID
		Attributes struct {
			ComDockerSwarmNodeID      string `json:"com.docker.swarm.node.id"`      // swarm节点ID
			ComDockerSwarmServiceID   string `json:"com.docker.swarm.service.id"`   // swarm服务ID
			ComDockerSwarmServiceName string `json:"com.docker.swarm.service.name"` // swarm服务名称（容器名称）
			ComDockerSwarmTask        string `json:"com.docker.swarm.task"`
			ComDockerSwarmTaskID      string `json:"com.docker.swarm.task.id"`
			ComDockerSwarmTaskName    string `json:"com.docker.swarm.task.name"` // 节点所在的实例名称
			Image                     string `json:"image"`                      // 镜像
			Name                      string `json:"name"`                       // 节点所在的实例名称
		} `json:"Attributes"`
	} `json:"Actor"`
	Scope    string `json:"scope"`
	Time     int    `json:"time"`     // 发生时间（秒）
	TimeNano int64  `json:"timeNano"` // 发生时间（纳秒）
}
