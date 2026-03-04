package docker

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"time"

	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
)

type service struct {
	unixClient *http.Client
}

// Inspect 查看服务详情
func (receiver service) Inspect(serviceName string) (ServiceInspectJson, error) {
	// 1. 构造 URL
	// 直接访问 /services/{name}，API 返回单个对象
	url := fmt.Sprintf("http://localhost/services/%s", serviceName)

	// 2. 调用工具函数解析
	// 因为 ServiceInspectJson 是结构体，直接对应 API 返回的 { ... }
	result, _ := UnixGetDecode[ServiceInspectJson](receiver.unixClient, url)

	// 3. 判断是否存在
	// 如果服务不存在 (404)，API 返回 {"message": "..."}，解析后结构体字段均为默认值
	// 通过判断关键字段 ID 是否为空来确定服务是否存在
	if result.ID == "" {
		return result, errors.New("no such service")
	}

	return result, nil
}

// Exists 服务是否存在
func (receiver service) Exists(serviceName string) bool {
	// 复用 Inspect 方法
	_, err := receiver.Inspect(serviceName)
	return err == nil
}

// Delete 删除容器服务
func (receiver service) Delete(serviceName string) error {
	// 1. 构造 URL
	// API: DELETE /services/{id}
	url := fmt.Sprintf("http://localhost/services/%s", serviceName)

	// 2. 调用工具函数 UnixDelete
	// UnixDelete 内部已处理了 204 判断和错误封装
	_, err := UnixDelete(receiver.unixClient, url)

	return err
}

// SetImagesAndReplicas 更新镜像版本和副本数量
func (receiver service) SetImagesAndReplicas(serviceName string, dockerImages string, dockerReplicas int) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker service update --image %s --replicas %v --with-registry-auth %s --update-order start-first", dockerImages, dockerReplicas, serviceName), nil, "", false)
}

// SetImages 更新镜像版本
func (receiver service) SetImages(serviceName string, dockerImages string, updateDelay int) (chan string, func() int) {
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("docker service update --image %s  --with-registry-auth --update-order start-first", dockerImages))

	// 滚动更新时的时间间隔
	if updateDelay > 0 {
		sb.WriteString(fmt.Sprintf(" --update-delay=%ds", updateDelay))
	}

	sb.WriteString(fmt.Sprintf(" %s", serviceName))

	return exec.RunShell(sb.String(), nil, "", true)
}

// SetReplicas 更新副本数量
func (receiver service) SetReplicas(serviceName string, dockerReplicas int) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker service update --replicas %v --with-registry-auth %s", dockerReplicas, serviceName), nil, "", false)
}

// Restart 重启容器
func (receiver service) Restart(serviceName string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker service update --with-registry-auth --force %s", serviceName), nil, "", false)
}

// Create 创建服务
func (receiver service) Create(serviceName, dockerNodeRole, additionalScripts, dockerNetwork string, dockerReplicas int, dockerImages string, limitCpus float64, limitMemory string) (chan string, func() int) {
	var sb bytes.Buffer
	sb.WriteString("docker service create --with-registry-auth --mount type=bind,src=/etc/localtime,dst=/etc/localtime")
	sb.WriteString(fmt.Sprintf(" --name %s -d --network=%s", serviceName, dockerNetwork))
	sb.WriteString(" --update-order start-first")

	// 节点筛选
	switch dockerNodeRole {
	case "global", "GLOBAL":
		sb.WriteString(" --mode global")
	case "":
		sb.WriteString(fmt.Sprintf("  --replicas %v", dockerReplicas))
	default:
		sb.WriteString(fmt.Sprintf("  --replicas %v --constraint node.role==%s", dockerReplicas, dockerNodeRole))
	}

	if limitCpus > 0 {
		sb.WriteString(fmt.Sprintf(" --limit-cpu=%f", limitCpus))
	}
	if limitMemory != "" {
		sb.WriteString(fmt.Sprintf(" --limit-memory=%s", limitMemory))
	}
	sb.WriteString(fmt.Sprintf(" %s %s", additionalScripts, dockerImages))

	return exec.RunShellContext(context.Background(), sb.String(), nil, "", true)
}

type ServiceLogVO struct {
	ContainerId string
	ServiceName string
	NodeName    string
	Logs        collections.List[string]
}

// Logs 获取日志
func (receiver service) Logs(serviceIdOrServiceName string, tailCount int) (collections.List[ServiceLogVO], error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/tasks?filters={"service":{"fops":true}}
	// docker service logs fops
	lstLog, exitCode := exec.RunShellCommand(fmt.Sprintf("docker service logs %s --tail %d", serviceIdOrServiceName, tailCount), nil, "", true)

	lst := collections.NewList[ServiceLogVO]()
	if exitCode != 0 {
		return lst, fmt.Errorf("获取日志失败。")
	}

	lstLog.Foreach(func(item *string) {
		logs := strings.SplitN(*item, "|", 2)
		if len(logs) != 2 {
			return
		}

		// 得到容器名称和节点名称
		name_Id_NodeName := strings.TrimSpace(logs[0])
		// 日志内容
		content := strings.TrimSpace(logs[1])
		// 节点名称
		name_Id, nodeName, _ := strings.Cut(name_Id_NodeName, "@")
		// 服务ID和名称
		var serverName string
		var containerId string
		if i := strings.LastIndex(name_Id, "."); i > 0 && i < len(name_Id)-1 {
			serverName, containerId = name_Id[:i], name_Id[i+1:]
		}

		// 找到对应的容器，添加日志
		if curContainer := lst.Find(func(item *ServiceLogVO) bool {
			return item.ContainerId == containerId
		}); curContainer != nil {
			curContainer.Logs.Add(content)
		} else {
			lst.Add(ServiceLogVO{
				ContainerId: containerId,
				ServiceName: serverName,
				NodeName:    nodeName,
				Logs:        collections.NewList(content),
			})
		}
	})
	return lst, nil
}

// ServiceListVO 容器的名称 实例数量 副本数量 镜像（docker service ls）
type ServiceListVO struct {
	ID   string `json:"ID"`
	Spec struct {
		Name         string                 `json:"Name"`
		Mode         map[string]interface{} `json:"Mode"` // 动态解析 Replicated 或 Global
		TaskTemplate struct {
			ContainerSpec struct {
				Image string `json:"Image"`
			} `json:"ContainerSpec"`
		} `json:"TaskTemplate"`
	} `json:"Spec"`
	Replicas int // 副本数量
}

// List 获取所有Service
func (receiver service) List() collections.List[ServiceListVO] {
	// 1. 获取服务列表
	// API: GET /services
	services, _ := UnixGetDecode[collections.List[ServiceListVO]](receiver.unixClient, "http://localhost/services")
	if services.Count() == 0 {
		return services
	}

	// 4. 组装数据
	services.Foreach(func(svc *ServiceListVO) {
		// 解析期望副本数
		// 尝试解析 Replicated 模式
		if repl, ok := svc.Spec.Mode["Replicated"].(map[string]interface{}); ok {
			if r, ok := repl["Replicas"].(float64); ok {
				svc.Replicas = int(r)
			}
		}
	})
	return services
}

// PS 获取容器运行的实例信息
func (receiver service) PS(lstNode collections.List[DockerNodeVO], serviceName string) collections.List[ServiceTaskVO] {
	lstTaskGroupVO := collections.NewList[ServiceTaskVO]()

	// 1. 获取任务列表
	// API 过滤器：{"service":{"serviceName":true}}
	filter := fmt.Sprintf(`{"service":{"%s":true}}`, serviceName)
	tasksUrl := "http://localhost/tasks?filters=" + url.QueryEscape(filter)

	type swarmTask struct {
		ID     string `json:"ID"`
		Slot   int    `json:"Slot"`
		NodeID string `json:"NodeID"`
		Status struct {
			Timestamp string `json:"Timestamp"`
			State     string `json:"State"`
			Err       string `json:"Err"`
		} `json:"Status"`
		Spec struct {
			ContainerSpec struct {
				Image string `json:"Image"`
			} `json:"ContainerSpec"`
		} `json:"Spec"`
		ServiceAnnotations struct {
			Name string `json:"Name"`
		} `json:"ServiceAnnotations"`
		// 全局服务可能需要 Index
		Index int `json:"Index"`
	}
	tasks, err := UnixGetDecode[[]swarmTask](receiver.unixClient, tasksUrl)
	if err != nil || len(tasks) == 0 {
		return lstTaskGroupVO
	}

	// 3. 按 Slot 分组
	// Key: Slot ID, Value: 任务列表
	slotMap := make(map[int][]swarmTask)
	for _, task := range tasks {
		// 排除全局服务（Slot为0）或 Slot 为 0 的情况，通常按 NodeID 分组
		// 这里假设大部分是副本服务，按 Slot 分组
		if task.Slot == 0 {
			// 简单处理：全局服务按 NodeID 归类，或者直接作为独立任务
			// 为防止 key 冲突，这里假设全局服务也要分组逻辑
			task.Slot = int(uintptr(task.Index)) // 简单防重，实际业务可能需要更复杂的逻辑
		}

		// 将节点ID转换成节点名称
		curNode := lstNode.Find(func(item *DockerNodeVO) bool {
			return item.ID == task.NodeID
		})
		if curNode != nil {
			task.NodeID = curNode.Description.Hostname
		}

		slotMap[task.Slot] = append(slotMap[task.Slot], task)
	}

	// 4. 处理分组数据
	for _, taskGroup := range slotMap {
		// 按时间排序：最新的排在前面
		sort.Slice(taskGroup, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339Nano, taskGroup[i].Status.Timestamp)
			t2, _ := time.Parse(time.RFC3339Nano, taskGroup[j].Status.Timestamp)
			return t1.After(t2)
		})

		// 最新的任务作为主任务
		if len(taskGroup) == 0 {
			continue
		}

		mainTask := taskGroup[0]
		// 构造 StateInfo (模拟 CLI 的 "Running 17 minutes ago")
		stateInfo := formatStateInfo(mainTask.Status.Timestamp, mainTask.Status.State)

		// 构造 VO
		vo := ServiceTaskVO{
			ServiceTaskId: mainTask.ID,
			Name:          fmt.Sprintf("%s.%d", mainTask.ServiceAnnotations.Name, mainTask.Slot), // 模拟 CLI 名称: service.1
			Image:         mainTask.Spec.ContainerSpec.Image,
			Node:          mainTask.NodeID, // 先放 NodeID，后续可以转换为 NodeName
			State:         mainTask.Status.State,
			StateInfo:     stateInfo,
			Error:         mainTask.Status.Err,
			Tasks:         collections.NewList[TaskInstanceVO](),
		}

		// 剩余任务作为子任务 (对应 CLI 中的 \_ )
		for i := 1; i < len(taskGroup); i++ {
			subTask := taskGroup[i]
			vo.Tasks.Add(TaskInstanceVO{
				TaskId:    subTask.ID,
				Image:     subTask.Spec.ContainerSpec.Image,
				Node:      subTask.NodeID,
				State:     subTask.Status.State,
				StateInfo: formatStateInfo(subTask.Status.Timestamp, subTask.Status.State),
				Error:     subTask.Status.Err,
			})
		}

		lstTaskGroupVO.Add(vo)
	}

	return lstTaskGroupVO
}

// formatStateInfo 简单的时间格式化函数
func formatStateInfo(timestamp, state string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return state
	}

	duration := time.Since(t)
	if duration.Seconds() < 60 {
		return fmt.Sprintf("%s %d seconds ago", state, int(duration.Seconds()))
	} else if duration.Minutes() < 60 {
		return fmt.Sprintf("%s %d minutes ago", state, int(duration.Minutes()))
	} else {
		return fmt.Sprintf("%s %d hours ago", state, int(duration.Hours()))
	}
}
