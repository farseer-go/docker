package docker

import (
	"bytes"
	"context"
	"errors"

	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type service struct {
	//progress chan string
}

// Delete 删除容器服务
func (receiver service) Delete(serviceName string) (chan string, func() int) {
	// docker service rm fops
	return exec.RunShell(fmt.Sprintf("docker service rm %s", serviceName), nil, "", false)
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

// Inspect 查看服务详情
func (receiver service) Inspect(serviceName string) (ServiceInspectJson, error) {
	// docker service inspect fops
	lst, _ := exec.RunShellCommand(fmt.Sprintf("docker service inspect %s", serviceName), nil, "", false)
	if lst.ContainsAny("no such service") {
		return nil, errors.New("no such service")
	}

	var serviceInspectJson ServiceInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := snc.Unmarshal([]byte(serviceInspectContent), &serviceInspectJson)

	return serviceInspectJson, err
}

// Exists 服务是否存在
func (receiver service) Exists(serviceName string) bool {
	serviceInspectJsons, err := receiver.Inspect(serviceName)
	if err != nil && strings.Contains(err.Error(), "no such service") {
		return false
	}
	if len(serviceInspectJsons) == 0 {
		return false
	}
	return serviceInspectJsons[0].ID != ""
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
	// docker service logs fops
	lstLog, exitCode := exec.RunShellCommand(fmt.Sprintf("docker service logs %s --tail %d", serviceIdOrServiceName, tailCount), nil, "", true)

	lst := collections.NewList[ServiceLogVO]()
	if exitCode != 0 {
		return lst, fmt.Errorf("获取日志失败。")
	}

	lstLog.Foreach(func(item *string) {
		logs := strings.SplitN(*item, "|", 2)
		if len(logs) != 2 {
			flog.Infof("容器日志分割|不成功：%s", *item)
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

// List 获取所有Service
func (receiver service) List() collections.List[ServiceListVO] {
	// docker service ls --format "table {{.ID}}|{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}"
	serviceList, exitCode := exec.RunShellCommand("docker service ls --format \"table {{.ID}}|{{.Name}}|{{.Replicas}}|{{.Image}}\"", nil, "", false)
	lstDockerName := collections.NewList[ServiceListVO]()
	if exitCode != 0 || serviceList.Count() == 0 {
		return lstDockerName
	}

	// 移除标题
	serviceList.RemoveAt(0)
	serviceList.Foreach(func(service *string) {
		// vwceboa7gtmu|redis|1/1|redis:latest
		sers := strings.Split(*service, "|")
		if len(sers) < 4 {
			return
		}
		insRepl := strings.Split(sers[2], "/")
		if len(insRepl) < 2 {
			insRepl = append(insRepl, "0")
		}
		lstDockerName.Add(ServiceListVO{
			Id:        sers[0],
			Name:      sers[1],
			Instances: parse.ToInt(insRepl[0]),
			Replicas:  parse.ToInt(insRepl[1]),
			Image:     sers[3],
		})
	})
	return lstDockerName
}

// PS 获取容器运行的实例信息
func (receiver service) PS(serviceName string) collections.List[ServiceTaskVO] {
	// docker service ps fops --format "table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}"
	serviceList, exitCode := exec.RunShellCommand(fmt.Sprintf("docker service ps %s --format \"table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}\"", serviceName), nil, "", false)
	lstTaskGroupVO := collections.NewList[ServiceTaskVO]()
	if exitCode != 0 || serviceList.Count() == 0 {
		return lstTaskGroupVO
	}

	// 移除标题
	serviceList.RemoveAt(0)
	serviceList.Foreach(func(service *string) {
		// whw9erkpysrj|fops|fops.552|test|Running|Running 17 minutes ago|
		sers := strings.Split(*service, "|")
		if len(sers) < 7 {
			return
		}
		// 包含\_的名称，说明是子任务
		if strings.Contains(sers[1], "\\_") {
			taskGroupVO := lstTaskGroupVO.LastAddr()
			taskGroupVO.Tasks.Add(TaskInstanceVO{
				TaskId:    sers[0],
				Image:     sers[2],
				Node:      sers[3],
				State:     sers[4],
				StateInfo: sers[5],
				Error:     sers[6],
			})
		} else {
			// 主任务
			lstTaskGroupVO.Add(ServiceTaskVO{
				ServiceTaskId: sers[0],
				Name:          sers[1],
				Image:         sers[2],
				Node:          sers[3],
				State:         sers[4],
				StateInfo:     sers[5],
				Error:         sers[6],
				Tasks:         collections.NewList[TaskInstanceVO](),
			})
		}
	})
	return lstTaskGroupVO
}
