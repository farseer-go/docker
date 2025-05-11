package docker

import (
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type event struct {
	progress chan string
}

// Watch 持续获取docker事件
func (receiver event) Watch(image string) chan EventResult {
	eventResultChan := make(chan EventResult, 1000)
	// 将读取到的json事件信息转换成EventResult结构体
	go func() {
		for json := range receiver.progress {
			var eventResult EventResult
			snc.Unmarshal([]byte(json), &eventResult)
			eventResultChan <- eventResult
		}
	}()
	go func() {
		exec.RunShell("docker events --filter 'type=container' --format '{{json .}}'", receiver.progress, nil, "", false)
		close(eventResultChan)
	}()
	return eventResultChan
}

// docker 事件结构体
type EventResult struct {
	Status string `json:"status"`
	ID     string `json:"id"`
	From   string `json:"from"`
	Type   string `json:"Type"`
	Action string `json:"Action"`
	Actor  struct {
		ID         string `json:"ID"`
		Attributes struct {
			ComDockerSwarmNodeID      string `json:"com.docker.swarm.node.id"`
			ComDockerSwarmServiceID   string `json:"com.docker.swarm.service.id"`
			ComDockerSwarmServiceName string `json:"com.docker.swarm.service.name"`
			ComDockerSwarmTask        string `json:"com.docker.swarm.task"`
			ComDockerSwarmTaskID      string `json:"com.docker.swarm.task.id"`
			ComDockerSwarmTaskName    string `json:"com.docker.swarm.task.name"`
			Image                     string `json:"image"`
			Name                      string `json:"name"`
		} `json:"Attributes"`
	} `json:"Actor"`
	Scope    string `json:"scope"`
	Time     int    `json:"time"`
	TimeNano int64  `json:"timeNano"`
}
