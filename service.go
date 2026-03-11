package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"time"

	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
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

	result.Spec.TaskTemplate.ContainerSpec.Image = strings.Split(result.Spec.TaskTemplate.ContainerSpec.Image, "@")[0]                 // 去掉 digest 部分
	result.PreviousSpec.TaskTemplate.ContainerSpec.Image = strings.Split(result.PreviousSpec.TaskTemplate.ContainerSpec.Image, "@")[0] // 去掉 digest 部分
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
func (receiver service) SetImagesAndReplicas(serviceName string, dockerImages string, dockerReplicas int) exec.ShellWait {
	args := []string{
		"service", "update",
		"--image", dockerImages,
		"--replicas", fmt.Sprintf("%d", dockerReplicas),
		"--with-registry-auth",
		"--update-order", "start-first",
		serviceName,
	}
	return exec.RunShell("docker", args, nil, "", false)
}

// SetImages 更新镜像版本
func (receiver service) SetImages(serviceName string, dockerImages string, updateDelay int) exec.ShellWait {
	args := []string{
		"service", "update",
		"--image", dockerImages,
		"--with-registry-auth",
		"--update-order", "start-first",
	}

	// 滚动更新时的时间间隔
	if updateDelay > 0 {
		args = append(args, fmt.Sprintf("--update-delay=%ds", updateDelay))
	}

	args = append(args, serviceName)

	return exec.RunShell("docker", args, nil, "", true)
}

// SetReplicas 更新副本数量
func (receiver service) SetReplicas(serviceName string, dockerReplicas int) exec.ShellWait {
	args := []string{
		"service", "update",
		"--replicas", fmt.Sprintf("%d", dockerReplicas),
		"--with-registry-auth",
		serviceName,
	}
	return exec.RunShell("docker", args, nil, "", false)
}

// Restart 重启容器
func (receiver service) Restart(serviceName string) exec.ShellWait {
	args := []string{
		"service", "update",
		"--with-registry-auth",
		"--force",
		serviceName,
	}
	return exec.RunShell("docker", args, nil, "", false)
}

type ConfigTarget struct {
	Name   string // 配置文件名称 docker config ls
	Target string // 挂载到容器内的文件
}

// Create 创建服务
func (receiver service) Create(serviceName, dockerNodeRole, additionalScripts, dockerNetwork string, dockerReplicas int, dockerImages string, limitCpus float64, limitMemory string, config ConfigTarget) exec.ShellWait {
	args := []string{
		"service", "create",
		"--with-registry-auth",
		"--mount", "type=bind,src=/etc/localtime,dst=/etc/localtime",
		"--name", serviceName,
		"-d",
		"--network=" + dockerNetwork,
		"--update-order", "start-first",
	}

	// 挂载 Docker Config（如果提供）
	if len(config.Name) > 0 && len(config.Target) > 0 {
		args = append(args, "--config", fmt.Sprintf("source=%s,target=%s", config.Name, config.Target))
	}

	// 节点筛选
	switch strings.ToUpper(dockerNodeRole) {
	case "GLOBAL":
		args = append(args, "--mode", "global")
	case "":
		args = append(args, "--replicas", fmt.Sprintf("%d", dockerReplicas))
	default:
		args = append(args, "--replicas", fmt.Sprintf("%d", dockerReplicas))
		args = append(args, "--constraint", "node.role=="+dockerNodeRole)
	}

	if limitCpus > 0 {
		args = append(args, fmt.Sprintf("--limit-cpu=%f", limitCpus))
	}
	if limitMemory != "" {
		args = append(args, "--limit-memory="+limitMemory)
	}

	// 额外参数
	if additionalScripts != "" {
		args = append(args, ParseShellArgs(additionalScripts)...)
	}

	args = append(args, dockerImages)

	return exec.RunShell("docker", args, nil, "", true)
}

// ParseShellArgs 解析 shell 风格的参数字符串（支持引号和续行符）
func ParseShellArgs(s string) []string {
	// 处理续行符：\ + 换行符 -> 空格
	s = strings.ReplaceAll(s, "\\\n", " ")
	s = strings.ReplaceAll(s, "\\\r\n", " ")
	// 清理多余空白
	s = strings.TrimSpace(s)

	var args []string
	var current strings.Builder
	var inQuote bool
	var quoteChar rune

	for _, ch := range s {
		switch {
		case ch == '"' || ch == '\'':
			if inQuote {
				if ch == quoteChar {
					inQuote = false
				} else {
					current.WriteRune(ch)
				}
			} else {
				inQuote = true
				quoteChar = ch
			}
		case (ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t') && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// Logs 获取日志
func (receiver service) Logs(serviceIdOrServiceName string, tailCount int) (collections.List[ServiceLogVO], error) {
	args := []string{
		"service", "logs",
		serviceIdOrServiceName,
		"--tail", fmt.Sprintf("%d", tailCount),
	}

	wait := exec.RunShell("docker", args, nil, "", true)
	lstLog, exitCode := wait.WaitToList()

	lst := collections.NewList[ServiceLogVO]()
	if exitCode != 0 {
		return lst, fmt.Errorf("获取日志失败。")
	}

	lstLog.Foreach(func(item *string) {
		// fops.1.l71hvj98bsqx@master    | 2026-03-04 17:57:02.672616 Initialization completed, total time: 344 ms
		logs := strings.SplitN(*item, "|", 2)
		if len(logs) != 2 {
			return
		}

		// 得到容器名称和节点名称 fops.1.l71hvj98bsqx@master
		name_Id_NodeName := strings.TrimSpace(logs[0])
		// 日志内容
		content := strings.TrimSpace(logs[1])
		// 节点名称 fops.1.l71hvj98bsqx		master
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

type ServiceLogVO struct {
	ContainerId string
	ServiceName string
	NodeName    string
	Logs        collections.List[string]
}

// ServiceListVO 容器的名称 实例数量 副本数量 镜像（docker service ls）
type ServiceListVO struct {
	ID   string `json:"ID"`
	Spec struct {
		Name string `json:"Name"`
		Mode struct {
			Replicated struct {
				Replicas int `json:"Replicas"` // 副本数量
			} `json:"Replicated,omitempty"`
			Global any
		} `json:"Mode"`
		TaskTemplate struct {
			ContainerSpec struct {
				Image   string              `json:"Image"`
				Configs []ServiceConfigJson `json:"Configs,omitempty"`
			} `json:"ContainerSpec"`
		} `json:"TaskTemplate"`
	} `json:"Spec"`
}

// List 获取所有Service
func (receiver service) List() collections.List[ServiceListVO] {
	// curl --unix-socket /var/run/docker.sock http://localhost/services
	// 1. 获取服务列表
	// API: GET /services
	services, _ := UnixGetDecode[collections.List[ServiceListVO]](receiver.unixClient, "http://localhost/services")
	if services.Count() == 0 {
		return services
	}

	// 4. 组装数据
	services.Foreach(func(svc *ServiceListVO) {
		svc.Spec.TaskTemplate.ContainerSpec.Image = strings.Split(svc.Spec.TaskTemplate.ContainerSpec.Image, "@")[0] // 去掉 digest 部分
	})
	return services
}

// PS 获取容器运行的实例信息
func (receiver service) PS(lstNode collections.List[DockerNodeVO], serviceName string) collections.List[ServiceTaskVO] {
	// curl --unix-socket /var/run/docker.sock http://localhost/tasks?filters=%7B%22service%22%3A%7B%22fops%22%3Atrue%7D%7D
	lstTaskGroupVO := collections.NewList[ServiceTaskVO]()

	// 1. 获取任务列表
	// API 过滤器：{"service":{"serviceName":true}}
	filter := fmt.Sprintf(`{"service":{"%s":true}}`, serviceName)
	tasksUrl := "http://localhost/tasks?filters=" + url.QueryEscape(filter)

	tasks, err := UnixGetDecode[[]ServiceIdInspectJson](receiver.unixClient, tasksUrl)
	if err != nil || len(tasks) == 0 {
		return lstTaskGroupVO
	}

	// 3. 按 Slot 分组
	// Key: Slot ID, Value: 任务列表
	slotMap := make(map[int][]ServiceIdInspectJson)
	for _, task := range tasks {
		// 排除全局服务（Slot为0）或 Slot 为 0 的情况，通常按 NodeID 分组
		// 这里假设大部分是副本服务，按 Slot 分组
		if task.Slot == 0 {
			// 简单处理：全局服务按 NodeID 归类，或者直接作为独立任务
			// 为防止 key 冲突，这里假设全局服务也要分组逻辑
			task.Slot = int(uintptr(task.Index)) // 简单防重，实际业务可能需要更复杂的逻辑
		}

		// 根据节点ID,找到对应的节点信息，补全节点名称和IP
		curNode := lstNode.Find(func(item *DockerNodeVO) bool {
			return item.ID == task.NodeID
		})
		if curNode != nil {
			task.NodeName = curNode.Description.Hostname
			task.NodeIP = curNode.Status.Addr
		}

		// 截断 ContainerID，保持和 CLI 输出一致
		if len(task.Status.ContainerStatus.ContainerID) >= 12 {
			task.Status.ContainerStatus.ContainerID = task.Status.ContainerStatus.ContainerID[:12]
		}

		task.Spec.ContainerSpec.Image = strings.Split(task.Spec.ContainerSpec.Image, "@")[0] // 去掉 digest 部分
		slotMap[task.Slot] = append(slotMap[task.Slot], task)
	}

	// 4. 处理分组数据
	// 辅助函数：定义状态优先级 (数字越小优先级越高，越应该排在前面)
	getStatePriority := func(state string) int {
		switch state {
		case "running", "starting", "pending", "ready":
			return 1 // 活跃状态优先级最高
		case "failed", "rejected":
			return 3 // 失败状态优先级最低
		case "shutdown", "complete", "orphaned":
			return 2 // 结束状态优先级中等
		default:
			return 4
		}
	}
	for _, taskGroup := range slotMap {
		// 按时间排序：最新的排在前面
		sort.Slice(taskGroup, func(i, j int) bool {
			// t1, _ := time.Parse(time.RFC3339Nano, taskGroup[i].Status.Timestamp)
			// t2, _ := time.Parse(time.RFC3339Nano, taskGroup[j].Status.Timestamp)
			// return t1.After(t2)
			// 1. 首先比较状态优先级 (优先展示 Running 的任务)
			pI := getStatePriority(taskGroup[i].Status.State)
			pJ := getStatePriority(taskGroup[j].Status.State)
			if pI != pJ {
				return pI < pJ // 优先级小的排前面
			}
			return taskGroup[i].CreatedAt.After(taskGroup[j].CreatedAt)
		})

		// 最新的任务作为主任务
		if len(taskGroup) == 0 {
			continue
		}

		mainTask := taskGroup[0]

		// 构造 VO
		mainVO := ServiceTaskVO{
			ServiceTaskId: mainTask.ID,
			Name:          fmt.Sprintf("%s.%d", serviceName, mainTask.Slot), // 模拟 CLI 名称: service.1
			Image:         mainTask.Spec.ContainerSpec.Image,
			NodeID:        mainTask.NodeID,
			NodeName:      mainTask.NodeName,
			NodeIP:        mainTask.NodeIP,
			State:         mainTask.Status.State,
			StateInfo:     formatStateInfo(mainTask.Status.Timestamp, mainTask.Status.State), // 模拟 CLI 的 "Running 17 minutes ago"
			Error:         mainTask.Status.Err,
			Tasks:         collections.NewList[TaskInstanceVO](),
			CreatedAt:     mainTask.CreatedAt,
			UpdatedAt:     mainTask.UpdatedAt,
		}
		if len(mainTask.NetworksAttachments) > 0 {
			mainVO.Addresses = mainTask.NetworksAttachments[0].Addresses
		}

		// 剩余任务作为子任务 (对应 CLI 中的 \_ )
		for i := 1; i < len(taskGroup); i++ {
			subTask := taskGroup[i]
			subVO := TaskInstanceVO{
				TaskId:    subTask.ID,
				Image:     subTask.Spec.ContainerSpec.Image,
				NodeID:    subTask.NodeID,
				NodeName:  subTask.NodeName,
				NodeIP:    subTask.NodeIP,
				State:     subTask.Status.State,
				StateInfo: formatStateInfo(subTask.Status.Timestamp, subTask.Status.State), // 模拟 CLI 的 "Running 17 minutes ago"
				Error:     subTask.Status.Err,
				CreatedAt: subTask.CreatedAt,
				UpdatedAt: subTask.UpdatedAt,
			}
			if len(subTask.NetworksAttachments) > 0 {
				subVO.Addresses = subTask.NetworksAttachments[0].Addresses
			}
			mainVO.Tasks.Add(subVO)
		}

		lstTaskGroupVO.Add(mainVO)
	}

	return lstTaskGroupVO
}

// formatStateInfo 简单的时间格式化函数
func formatStateInfo(timestamp time.Time, state string) string {
	duration := time.Since(timestamp)
	if duration.Seconds() < 60 {
		return fmt.Sprintf("%s %d seconds ago", state, int(duration.Seconds()))
	} else if duration.Minutes() < 60 {
		return fmt.Sprintf("%s %d minutes ago", state, int(duration.Minutes()))
	} else {
		return fmt.Sprintf("%s %d hours ago", state, int(duration.Hours()))
	}
}

func (receiver service) UpdateServiceConfig(serviceName string, newConfigID, newConfigName, targetPath string) (bool, error) {
	// 1. 获取原始 JSON 到 map 中
	url := fmt.Sprintf("http://localhost/services/%s", serviceName)
	resp, _ := receiver.unixClient.Get(url)
	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)
	resp.Body.Close()

	// 2. 提取 Spec 和 Version
	spec := raw["Spec"].(map[string]interface{})
	version := int(raw["Version"].(map[string]interface{})["Index"].(float64))

	// 3. 动态修改 Configs 数组
	taskTemplate := spec["TaskTemplate"].(map[string]interface{})
	containerSpec := taskTemplate["ContainerSpec"].(map[string]interface{})
	configs := containerSpec["Configs"].([]interface{})

	for i := range configs {
		cfg := configs[i].(map[string]interface{})
		// 匹配 targetPath
		file := cfg["File"].(map[string]interface{})
		if file["Name"] == targetPath {
			cfg["ConfigID"] = newConfigID
			cfg["ConfigName"] = newConfigName
			// 保留原有权限，只改 ID
			break
		}
	}

	// 4. 发送更新
	updateURL := fmt.Sprintf("http://localhost/services/%s/update?version=%d", serviceName, version)
	body, _ := json.Marshal(spec) // 只发 Spec 部分

	req, _ := http.NewRequest("POST", updateURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := receiver.unixClient.Do(req)
	resp.Body.Close()
	return resp.StatusCode == 200, err
}

// Inspect 查看服务详情
func (receiver service) GetCurConfigVersion(serviceName string) (int, error) {
	result, err := receiver.Inspect(serviceName)
	for _, config := range result.Spec.TaskTemplate.ContainerSpec.Configs {
		// 构建匹配格式，例如 "fops_config_v%d"
		format := fmt.Sprintf("%s_config_v", serviceName)
		if len(config.ConfigName) > len(format) {
			return parse.ToInt(config.ConfigName[len(format):]), nil
		}
	}

	return 0, err
}
