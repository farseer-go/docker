package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
	"path"
)

type container struct {
}

// Exists 判断容器是否已创建
func (receiver container) Exists(containerId string) bool {
	progress := make(chan string, 1000)
	// docker inspect fops
	var exitCode = exec.RunShell(fmt.Sprintf("docker inspect %s", containerId), progress, nil, "", false)
	lst := collections.NewListFromChan(progress)
	if exitCode != 0 {
		if lst.Contains("[]") && lst.ContainsPrefix("Error: No such object:") {
			return false
		}
		return false
	}
	if lst.Contains("[]") && lst.ContainsPrefix("Error: No such object:") {
		return false
	}
	return lst.ContainsAny(fmt.Sprintf("\"Name\": \"/%s\",", containerId))
}

// Kill 停止容器并删除
func (receiver container) Kill(containerId string) {
	exec.RunShell(fmt.Sprintf("docker kill %s", containerId), make(chan string, 1000), nil, "", false)
}

// RM 删除容器
func (receiver container) RM(containerId string) {
	exec.RunShell(fmt.Sprintf("docker rm %s", containerId), make(chan string, 1000), nil, "", false)
}

func (receiver container) Run(containerId string, networkName string, dockerImage string, args []string, useRm bool, env map[string]string, ctx context.Context) error {
	bf := bytes.Buffer{}
	bf.WriteString("docker run")
	if useRm {
		bf.WriteString(" --rm")
	}
	if containerId != "" {
		bf.WriteString(" --name ")
		bf.WriteString(containerId)
	}
	if networkName != "" {
		bf.WriteString(" --network=")
		bf.WriteString(networkName)
	}

	if args != nil {
		for _, arg := range args {
			bf.WriteString(" " + arg)
		}
	}

	bf.WriteString(" ")
	bf.WriteString(dockerImage)

	c := make(chan string, 100)
	exitCode := exec.RunShellContext(ctx, bf.String(), c, env, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

func (receiver container) Exec(containerId string, execCmd string, env map[string]string, progress chan string, ctx context.Context) error {
	if env == nil {
		env = make(map[string]string)
	}
	env["BASH_ENV"] = "\"/root/.bashrc\""

	bf := bytes.Buffer{}
	bf.WriteString("docker exec ") // docker exec FOPS-Build-hub-fsgit-cc-fops-130 echo aaa
	for k, v := range env {
		bf.WriteString(fmt.Sprintf("-e %s=%s ", k, v))
	}
	bf.WriteString(containerId)
	bf.WriteString(" /bin/bash -c ") //x
	bf.WriteString("\"")
	bf.WriteString(execCmd)
	bf.WriteString("\"")
	exitCode := exec.RunShellContext(ctx, bf.String(), progress, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf("docker exec 执行失败")
	}
	return nil
}

// Cp 复制文件到容器内
func (receiver container) Cp(containerId string, sourceFile, destFile string, ctx context.Context) error {
	c := make(chan string, 100)
	_ = receiver.Exec(containerId, "mkdir -p "+path.Dir(destFile), nil, c, ctx)

	// docker cp /var/lib/fops/dist/Dockerfile FOPS-Build:/var/lib/fops/dist/Dockerfile
	bf := bytes.Buffer{}
	bf.WriteString("docker cp ")
	bf.WriteString(sourceFile)
	bf.WriteString(" ")
	bf.WriteString(containerId)
	bf.WriteString(":")
	bf.WriteString(destFile)
	exitCode := exec.RunShellContext(ctx, bf.String(), c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// Logs 获取日志
func (receiver container) Logs(containerId string, tailCount int) collections.List[string] {
	c := make(chan string, 1000)
	// docker service logs fops
	exitCode := exec.RunShell(fmt.Sprintf("docker logs %s --tail %d", containerId, tailCount), c, nil, "", true)
	lst := collections.NewListFromChan(c)
	if exitCode != 0 {
		lst.Insert(0, "获取日志失败。")
	}
	return lst
}

// Inspect 查看容器详情
func (receiver container) Inspect(containerId string) (ContainerIdInspectJson, error) {
	progress := make(chan string, 1000)
	// docker inspect rqcinkiry0jr
	exec.RunShell(fmt.Sprintf("docker inspect %s", containerId), progress, nil, "", false)
	lst := collections.NewListFromChan(progress)
	if lst.ContainsAny("No such object") {
		return nil, nil
	}

	var containerInspectJson ContainerIdInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := json.Unmarshal([]byte(serviceInspectContent), &containerInspectJson)

	return containerInspectJson, err
}

// InspectByServiceId 查看服务详情
func (receiver container) InspectByServiceId(serviceId string) (ServiceIdInspectJson, error) {
	progress := make(chan string, 1000)
	// docker inspect rqcinkiry0jr
	exec.RunShell(fmt.Sprintf("docker inspect %s", serviceId), progress, nil, "", false)
	lst := collections.NewListFromChan(progress)
	if lst.ContainsAny("No such object") {
		return nil, nil
	}

	var serviceIdInspectJson ServiceIdInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := json.Unmarshal([]byte(serviceInspectContent), &serviceIdInspectJson)

	// 使用简短的容器ID
	if len(serviceIdInspectJson[0].Status.ContainerStatus.ContainerID) >= 12 {
		serviceIdInspectJson[0].Status.ContainerStatus.ContainerID = serviceIdInspectJson[0].Status.ContainerStatus.ContainerID[:12]
	}
	return serviceIdInspectJson, err
}
