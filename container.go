package docker

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"fmt"
	"path"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

type container struct {
	unixClient *http.Client
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

// Container 容器信息
type Container struct {
	ID      string        `json:"Id"`
	Names   []string      `json:"Names"`
	Image   string        `json:"Image"`
	ImageID string        `json:"ImageID"`
	Command string        `json:"Command"`
	Created int           `json:"Created"`
	Ports   []interface{} `json:"Ports"`
	Labels  struct {
		ComDockerSwarmNodeID      string `json:"com.docker.swarm.node.id"`
		ComDockerSwarmServiceID   string `json:"com.docker.swarm.service.id"`
		ComDockerSwarmServiceName string `json:"com.docker.swarm.service.name"`
		ComDockerSwarmTask        string `json:"com.docker.swarm.task"`
		ComDockerSwarmTaskID      string `json:"com.docker.swarm.task.id"`
		ComDockerSwarmTaskName    string `json:"com.docker.swarm.task.name"`
	} `json:"Labels"`
	State      string `json:"State"`
	Status     string `json:"Status"`
	HostConfig struct {
		NetworkMode string `json:"NetworkMode"`
	} `json:"HostConfig"`
	NetworkSettings struct {
		Networks struct {
			Net struct {
				IPAMConfig struct {
					IPv4Address string `json:"IPv4Address"`
				} `json:"IPAMConfig"`
				Links               interface{} `json:"Links"`
				Aliases             interface{} `json:"Aliases"`
				MacAddress          string      `json:"MacAddress"`
				NetworkID           string      `json:"NetworkID"`
				EndpointID          string      `json:"EndpointID"`
				Gateway             string      `json:"Gateway"`
				IPAddress           string      `json:"IPAddress"`
				IPPrefixLen         int         `json:"IPPrefixLen"`
				IPv6Gateway         string      `json:"IPv6Gateway"`
				GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
				GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
				DriverOpts          interface{} `json:"DriverOpts"`
				DNSNames            interface{} `json:"DNSNames"`
			} `json:"net"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	Mounts []struct {
		Type        string `json:"Type"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
		Propagation string `json:"Propagation"`
		Name        string `json:"Name,omitempty"`
		Driver      string `json:"Driver,omitempty"`
	} `json:"Mounts"`
}

// List 获取容器列表
func (receiver container) List(status string, labels map[string]string) (collections.List[Container], error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/containers/json?status=
	url := "http://localhost/containers/json?status=" + status
	for k, v := range labels {
		url += "&label=" + k + "=" + v
	}

	containers, err := UnixGet[collections.List[Container]](receiver.unixClient, url)
	return containers, err
}

// 解析响应
type StatsResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"` // 当前累计 CPU 使用时间（纳秒）
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_cpu_usage"` // 当前系统累计 CPU 时间（纳秒）
		OnlineCPUs  uint64 `json:"online_cpus"`      // CPU 核心数
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"` // 上一次累计 CPU 使用时间（纳秒）
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_cpu_usage"` // 上一次系统累计 CPU 时间（纳秒）
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
		Stats struct {
			Cache        uint64 `json:"cache"`
			RSS          uint64 `json:"rss"`
			InactiveFile uint64 `json:"inactive_file"` // 关键字段
		} `json:"stats"`
	} `json:"memory_stats"`
}

// getContainerStats 获取单个容器的统计信息
// names 参数是容器的名称列表，通常来自 List 方法的结果。Docker Swarm 模式下，容器名称格式为 /服务名.序号.任务ID
func (receiver container) Stats(containerID string) DockerStatsVO {
	dockerStatsVO := DockerStatsVO{
		ContainerID: containerID[:12],
	}

	// curl --unix-socket /var/run/docker.sock http://localhost/containers/9e76ea4b0231/stats?stream=false
	url := fmt.Sprintf("http://localhost/containers/%s/stats?stream=false", containerID)

	stats, err := UnixGet[StatsResponse](receiver.unixClient, url)
	if err != nil {
		return dockerStatsVO
	}

	// 解析容器名称（Swarm 格式: /服务名.序号.任务ID）
	// stats.Name = "/fops.1.l7c3377cnjacuy9xtz88resrw"
	// 移除前导斜杠
	containerName := strings.TrimPrefix(stats.Name, "/")
	parts := strings.Split(containerName, ".")
	// 补齐到 3 个部分
	for len(parts) < 3 {
		parts = append(parts, "")
	}

	dockerStatsVO.ContainerName = parts[0] + "." + parts[1]
	dockerStatsVO.Name = parts[0]
	dockerStatsVO.TaskId = parts[2]

	// taskId 最多取 12 位
	if len(dockerStatsVO.TaskId) > 12 {
		dockerStatsVO.TaskId = dockerStatsVO.TaskId[:12]
	}

	// 计算 CPU 使用率
	// 容器在这两次采样之间实际使用了多少 CPU 时间。
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	// 系统所有 CPU 核心在这两次采样之间的总可用时间。
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	// 容器在 4 核 CPU 上使用了 200%（相当于占用了 2 个核心）
	if systemDelta > 0 {
		dockerStatsVO.CpuUsagePercent = (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
	}

	// 计算内存使用率（MB）
	memUsage := stats.MemoryStats.Usage
	if stats.MemoryStats.Stats.InactiveFile > 0 {
		// 优先使用 inactive_file（与 Docker CLI 源码一致）
		memUsage -= stats.MemoryStats.Stats.InactiveFile
	} else if stats.MemoryStats.Stats.Cache > 0 {
		// 没有 inactive_file 时使用 cache
		memUsage -= stats.MemoryStats.Stats.Cache
	}

	dockerStatsVO.MemoryUsage = memUsage / 1024 / 1024
	dockerStatsVO.MemoryLimit = stats.MemoryStats.Limit / 1024 / 1024
	if stats.MemoryStats.Limit > 0 {
		dockerStatsVO.MemoryUsagePercent = float64(memUsage) / float64(stats.MemoryStats.Limit) * 100
	}

	return dockerStatsVO
}
