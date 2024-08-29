package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/utils/exec"
	"strings"
)

type service struct {
	dockerClient *client.Client
}

// Delete 删除容器服务
func (receiver service) Delete(containerId string) error {
	// docker service rm fops
	c := make(chan string, 1000)
	var exitCode = exec.RunShell(fmt.Sprintf("docker service rm %s", containerId), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// SetImagesAndReplicas 更新镜像版本和副本数量
func (receiver service) SetImagesAndReplicas(containerId string, dockerImages string, dockerReplicas int) error {
	c := make(chan string, 1000)
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --image %s --replicas %v --update-delay 10s --with-registry-auth %s", dockerImages, dockerReplicas, containerId), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// SetImages 更新镜像版本
func (receiver service) SetImages(containerId string, dockerImages string) error {
	c := make(chan string, 1000)
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --image %s --update-delay 10s --with-registry-auth %s", dockerImages, containerId), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// SetReplicas 更新副本数量
func (receiver service) SetReplicas(containerId string, dockerReplicas int) error {
	c := make(chan string, 1000)
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --replicas %v --with-registry-auth %s", dockerReplicas, containerId), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// Restart 重启容器
func (receiver service) Restart(containerId string) error {
	c := make(chan string, 1000)
	var exitCode = exec.RunShell(fmt.Sprintf("docker service update --with-registry-auth --force %s", containerId), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// Inspect 查看服务详情
func (receiver service) Inspect(containerId string) (ServiceInspectJson, error) {
	progress := make(chan string, 1000)
	// docker service inspect fops
	exec.RunShell(fmt.Sprintf("docker service inspect %s", containerId), progress, nil, "", false)
	lst := collections.NewListFromChan(progress)
	if lst.ContainsAny("no such service") {
		return nil, nil
	}

	var serviceInspectJson ServiceInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := json.Unmarshal([]byte(serviceInspectContent), &serviceInspectJson)

	return serviceInspectJson, err
}

// Exists 服务是否存在
func (receiver service) Exists(containerId string) (bool, error) {
	serviceInspectJsons, err := receiver.Inspect(containerId)
	if err != nil && strings.Contains(err.Error(), " not found") {
		return false, nil
	}
	if len(serviceInspectJsons) == 0 {
		return false, err
	}
	return serviceInspectJsons[0].ID != "", err
}

// Create 创建服务
func (receiver service) Create(containerId, dockerNodeRole, additionalScripts, dockerNetwork string, dockerReplicas int, dockerImages string, limitCpus float64, limitMemory string) error {
	c := make(chan string, 1000)
	var sb bytes.Buffer
	sb.WriteString("docker service create --with-registry-auth --mount type=bind,src=/etc/localtime,dst=/etc/localtime")
	sb.WriteString(fmt.Sprintf(" --name %s -d --network=%s", containerId, dockerNetwork))

	// 所有节点都要运行
	if dockerNodeRole == "global" {
		sb.WriteString(" --mode global")
	} else {
		sb.WriteString(fmt.Sprintf("  --replicas %v --constraint node.role==%s", dockerReplicas, dockerNodeRole))
	}

	if limitCpus > 0 {
		sb.WriteString(fmt.Sprintf(" --limit-cpu=%f", limitCpus))
	}
	if limitMemory != "" {
		sb.WriteString(fmt.Sprintf(" --limit-memory=%s", limitMemory))
	}
	sb.WriteString(fmt.Sprintf(" %s %s", additionalScripts, dockerImages))

	var exitCode = exec.RunShellContext(context.Background(), sb.String(), c, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// Logs 获取日志
func (receiver service) Logs(containerId string, tailCount int) (collections.List[string], error) {
	progress := make(chan string, 1000)
	// docker service logs fops
	var exitCode = exec.RunShell(fmt.Sprintf("docker service logs %s --tail %d", containerId, tailCount), progress, nil, "", true)
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
	progress := make(chan string, 1000)
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
		lstDockerName.Add(ServiceListVO{
			Id:        sers[0],
			Name:      sers[1],
			Instances: parse.ToInt(strings.Split(sers[2], "/")[0]),
			Replicas:  parse.ToInt(strings.Split(sers[2], "/")[1]),
			Image:     sers[3],
		})
	})
	return lstDockerName
}

// PS 获取容器运行的实例信息
func (receiver service) PS(containerId string) collections.List[ServicePsVO] {
	progress := make(chan string, 1000)
	// docker service ps fops --format "table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}"
	var exitCode = exec.RunShell(fmt.Sprintf("docker service ps %s --format \"table {{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}\"", containerId), progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
	lstDockerInstance := collections.NewList[ServicePsVO]()
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
		lstDockerInstance.Add(ServicePsVO{
			Id:        sers[0],
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
