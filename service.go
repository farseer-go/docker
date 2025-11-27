package docker

import (
	"bytes"
	"context"

	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type service struct {
	progress chan string
}

// Delete 删除容器服务
func (receiver service) Delete(serviceName string) error {
	// docker service rm fops
	var exitCode = exec.RunShell(fmt.Sprintf("docker service rm %s", serviceName), receiver.progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// SetImagesAndReplicas 更新镜像版本和副本数量
func (receiver service) SetImagesAndReplicas(serviceName string, dockerImages string, dockerReplicas int) error {
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --image %s --replicas %v --with-registry-auth %s", dockerImages, dockerReplicas, serviceName), receiver.progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// SetImages 更新镜像版本
func (receiver service) SetImages(serviceName string, dockerImages string, updateDelay int) error {
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("docker service update --image %s  --with-registry-auth", dockerImages))

	// 滚动更新时的时间间隔
	if updateDelay > 0 {
		sb.WriteString(fmt.Sprintf(" --update-delay=%ds", updateDelay))
	}

	sb.WriteString(fmt.Sprintf(" %s", serviceName))

	var exitCode = exec.RunShell(sb.String(), receiver.progress, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// SetReplicas 更新副本数量
func (receiver service) SetReplicas(serviceName string, dockerReplicas int) error {
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --replicas %v --with-registry-auth %s", dockerReplicas, serviceName), receiver.progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// Restart 重启容器
func (receiver service) Restart(serviceName string) error {
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --with-registry-auth --force %s", serviceName), receiver.progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// Inspect 查看服务详情
func (receiver service) Inspect(serviceName string) (ServiceInspectJson, error) {
	progress := make(chan string, 5000)
	// docker service inspect fops
	exec.RunShell(fmt.Sprintf("docker service inspect %s", serviceName), progress, nil, "", false)
	lst := collections.NewListFromChan(progress)
	if lst.ContainsAny("no such service") {
		return nil, nil
	}

	var serviceInspectJson ServiceInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := snc.Unmarshal([]byte(serviceInspectContent), &serviceInspectJson)

	return serviceInspectJson, err
}

// Exists 服务是否存在
func (receiver service) Exists(serviceName string) (bool, error) {
	serviceInspectJsons, err := receiver.Inspect(serviceName)
	if err != nil && strings.Contains(err.Error(), " not found") {
		return false, nil
	}
	if len(serviceInspectJsons) == 0 {
		return false, err
	}
	return serviceInspectJsons[0].ID != "", err
}

// Create 创建服务
func (receiver service) Create(serviceName, dockerNodeRole, additionalScripts, dockerNetwork string, dockerReplicas int, dockerImages string, limitCpus float64, limitMemory string) error {
	var sb bytes.Buffer
	sb.WriteString("docker service create --with-registry-auth --mount type=bind,src=/etc/localtime,dst=/etc/localtime")
	sb.WriteString(fmt.Sprintf(" --name %s -d --network=%s", serviceName, dockerNetwork))

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

	var exitCode = exec.RunShellContext(context.Background(), sb.String(), receiver.progress, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
	}
	return nil
}

// Logs 获取日志
func (receiver service) Logs(serviceIdOrServiceName string, tailCount int) (collections.List[string], error) {
	progress := make(chan string, 5000)
	// docker service logs fops
	var exitCode = exec.RunShell(fmt.Sprintf("docker service logs %s --tail %d", serviceIdOrServiceName, tailCount), progress, nil, "", true)
	lst := collections.NewListFromChan(progress)
	if exitCode != 0 {
		return lst, fmt.Errorf("获取日志失败。")
	}
	lst.Foreach(func(item *string) {
		if strings.Contains(*item, "|") {
			*item = strings.TrimSpace(strings.SplitN(*item, "|", 2)[1])
		}
	})
	return lst, nil
}

// List 获取所有Service
func (receiver service) List() collections.List[ServiceListVO] {
	progress := make(chan string, 5000)
	// docker service ls --format "table {{.ID}}|{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}"
	var exitCode = exec.RunShell("docker service ls --format \"table {{.ID}}|{{.Name}}|{{.Replicas}}|{{.Image}}\"", progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
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
func (receiver service) PS(serviceName string) collections.List[TaskInstanceVO] {
	progress := make(chan string, 5000)
	// docker service ps fops --format "table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}"
	var exitCode = exec.RunShell(fmt.Sprintf("docker service ps %s --format \"table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}\"", serviceName), progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
	lstDockerInstance := collections.NewList[TaskInstanceVO]()
	if exitCode != 0 || serviceList.Count() == 0 {
		return lstDockerInstance
	}

	// 移除标题
	serviceList.RemoveAt(0)
	serviceList.Foreach(func(service *string) {
		// whw9erkpysrj|fops|fops.552|test|Running|Running 17 minutes ago|
		sers := strings.Split(*service, "|")
		if len(sers) < 7 {
			return
		}
		lstDockerInstance.Add(TaskInstanceVO{
			TaskId:    sers[0],
			Name:      sers[1],
			Image:     sers[2],
			Node:      sers[3],
			State:     sers[4],
			StateInfo: sers[5],
			Error:     sers[6],
		})
	})
	return lstDockerInstance
}
