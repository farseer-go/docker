package docker

import (
	"net/http"
	"sync"

	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type event struct {
	unixClient *http.Client
	handlers   []EventHandler
	mu         sync.RWMutex
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(event EventResult)
}

// EventHandlerFunc 函数类型适配器
type EventHandlerFunc func(event EventResult)

// Handle 实现 EventHandler 接口
func (f EventHandlerFunc) Handle(event EventResult) {
	f(event)
}

// EventResult docker 事件结构体
type EventResult struct {
	Status string `json:"status"` // 事件类型
	ID     string `json:"id"`     // 容器ID
	From   string `json:"from"`   // 镜像
	Type   string `json:"Type"`   // container
	Action string `json:"Action"` // start
	Actor  struct {
		ID         string `json:"ID"` // 容器ID
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

// Register 注册事件处理器
func (receiver *event) Register(handler EventHandler) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.handlers = append(receiver.handlers, handler)
}

// RegisterFunc 注册事件处理函数（便捷方法）
func (receiver *event) RegisterFunc(handler func(event EventResult)) {
	receiver.Register(EventHandlerFunc(handler))
}

// Start 启动事件监听（只启动一次）
func (receiver *event) Start() {
	wait := exec.RunShell("docker", []string{"events", "--format", "{{json .}}"}, nil, "", false)
	go wait.WaitToFunc(func(json string) {
		var eventResult EventResult
		snc.Unmarshal([]byte(json), &eventResult)

		// 广播给所有处理器
		receiver.mu.RLock()
		handlers := make([]EventHandler, len(receiver.handlers))
		copy(handlers, receiver.handlers)
		receiver.mu.RUnlock()

		for _, h := range handlers {
			h.Handle(eventResult)
		}
	})
}
