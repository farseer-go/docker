package docker

import (
	"bytes"
	"context"

	"fmt"
	"path"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type container struct {
	//progress chan string
}

// Exists 判断容器是否已创建
func (receiver container) Exists(containerId string) bool {
	// docker inspect fops
	lstMessage, exitCode := exec.RunShellCommand(fmt.Sprintf("docker inspect %s", containerId), nil, "", false)
	if exitCode != 0 {
		if lstMessage.Contains("[]") && lstMessage.ContainsPrefix("Error: No such object:") {
			return false
		}
		return false
	}
	if lstMessage.Contains("[]") && lstMessage.ContainsPrefix("Error: No such object:") {
		return false
	}
	return lstMessage.ContainsAny(fmt.Sprintf("\"Name\": \"/%s\",", containerId))
}

// Kill 停止容器并删除
func (receiver container) Kill(containerId string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker kill %s", containerId), nil, "", false)
}

// RM 删除容器
func (receiver container) RM(containerId string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker rm %s", containerId), nil, "", false)
}

// 运行容器
func (receiver container) Run(containerId string, networkName string, dockerImage string, args []string, useRm bool, env map[string]string, ctx context.Context) (chan string, func() int) {
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

	return exec.RunShellContext(ctx, bf.String(), env, "", true)
}

// Restart 重启容器
func (receiver container) Restart(containerId string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker restart %s", containerId), nil, "", false)
}

// 在容器内部执行cmd命令
func (receiver container) Exec(containerId string, execCmd string, env map[string]string, ctx context.Context) (chan string, func() int) {
	if env == nil {
		env = make(map[string]string)
	}
	env["BASH_ENV"] = "\"/root/.bashrc\""

	bf := bytes.Buffer{}
	bf.WriteString("docker exec ") // docker exec FOPS-Build /bin/bash -c "xxxx.sh"
	for k, v := range env {
		bf.WriteString(fmt.Sprintf("-e %s=%s ", k, v))
	}
	bf.WriteString(containerId)
	bf.WriteString(" /bin/bash -c ") //x
	bf.WriteString("\"")
	bf.WriteString(execCmd)
	bf.WriteString("\"")
	return exec.RunShellContext(ctx, bf.String(), nil, "", false)
}

// Cp 复制文件到容器内
func (receiver container) Cp(containerId string, sourceFile, destFile string, ctx context.Context) (chan string, func() int) {
	_, wait := receiver.Exec(containerId, "mkdir -p "+path.Dir(destFile), nil, ctx)
	wait()

	// docker cp /var/lib/fops/dist/Dockerfile FOPS-Build:/var/lib/fops/dist/Dockerfile
	bf := bytes.Buffer{}
	bf.WriteString("docker cp ")
	bf.WriteString(sourceFile)
	bf.WriteString(" ")
	bf.WriteString(containerId)
	bf.WriteString(":")
	bf.WriteString(destFile)
	return exec.RunShellContext(ctx, bf.String(), nil, "", false)
}

// Logs 获取日志
func (receiver container) Logs(containerId string, tailCount int) collections.List[string] {
	// docker service logs fops
	lst, exitCode := exec.RunShellCommand(fmt.Sprintf("docker logs %s --tail %d", containerId, tailCount), nil, "", true)
	if exitCode != 0 {
		lst.Insert(0, "获取日志失败。")
	}
	return lst
}

// Inspect 查看容器详情
func (receiver container) Inspect(containerId string) (ContainerIdInspectJson, error) {
	// docker inspect rqcinkiry0jr
	lst, _ := exec.RunShellCommand(fmt.Sprintf("docker inspect %s", containerId), nil, "", false)
	if lst.ContainsAny("No such object") {
		return nil, nil
	}

	var containerInspectJson ContainerIdInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := snc.Unmarshal([]byte(serviceInspectContent), &containerInspectJson)

	return containerInspectJson, err
}

// InspectByServiceId 查看服务详情
func (receiver container) InspectByServiceId(serviceId string) (ServiceIdInspectJson, error) {
	// docker inspect rqcinkiry0jr
	lst, _ := exec.RunShellCommand(fmt.Sprintf("docker inspect %s", serviceId), nil, "", false)
	if lst.ContainsAny("No such object") {
		return nil, nil
	}

	var serviceIdInspectJson ServiceIdInspectJson
	serviceInspectContent := lst.ToString("\n")
	err := snc.Unmarshal([]byte(serviceInspectContent), &serviceIdInspectJson)
	// 使用简短的容器ID
	if len(serviceIdInspectJson) > 0 && len(serviceIdInspectJson[0].Status.ContainerStatus.ContainerID) >= 12 {
		serviceIdInspectJson[0].Status.ContainerStatus.ContainerID = serviceIdInspectJson[0].Status.ContainerStatus.ContainerID[:12]
	}
	return serviceIdInspectJson, err
}
