package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"fmt"
	"path"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/utils/exec"
)

type container struct {
	unixClient *http.Client
}

// Exists 判断容器是否已创建
func (receiver container) Exists(containerId string) bool {
	url := fmt.Sprintf("http://localhost/containers/%s/json", containerId)

	// 使用 UnixGet (工具方法内已处理 Body 关闭)
	resp, err := UnixGet(receiver.unixClient, url)
	if err != nil {
		return false
	}

	return resp.StatusCode == http.StatusOK
}

// Kill 停止容器并删除
func (receiver container) Kill(containerId string) error {
	url := fmt.Sprintf("http://localhost/containers/%s/kill", containerId)

	// 使用 UnixPost，内部已判断 204 并生成 error
	_, err := UnixPost(receiver.unixClient, url)
	return err
}

// RM 删除容器
func (receiver container) RM(containerId string) error {
	url := fmt.Sprintf("http://localhost/containers/%s", containerId)

	// 使用 UnixDelete
	_, err := UnixDelete(receiver.unixClient, url)
	return err
}

// Restart 重启容器
func (receiver container) Restart(containerId string) error {
	url := fmt.Sprintf("http://localhost/containers/%s/restart", containerId)

	// 使用 UnixPost，内部已处理状态码判断和错误封装
	_, err := UnixPost(receiver.unixClient, url)
	return err
}

// 运行容器(使用Docker CLI客户端)
func (receiver container) Run(containerId string, networkName string, dockerImage string, args []string, useRm bool, env map[string]string, ctx context.Context) exec.ShellWait {
	// 构建 args
	dockerArgs := []string{"run"}

	if useRm {
		dockerArgs = append(dockerArgs, "--rm")
	}
	if containerId != "" {
		dockerArgs = append(dockerArgs, "--name", containerId)
	}
	if networkName != "" {
		dockerArgs = append(dockerArgs, "--network="+networkName)
	}
	if args != nil {
		dockerArgs = append(dockerArgs, args...)
	}

	dockerArgs = append(dockerArgs, dockerImage)

	return exec.RunShellContext(ctx, "docker", dockerArgs, env, "", true)
}

// 在容器内部执行cmd命令(使用Docker CLI客户端)
func (receiver container) Exec(containerId string, execCmd string, env map[string]string, ctx context.Context) exec.ShellWait {
	if env == nil {
		env = make(map[string]string)
	}
	//env["BASH_ENV"] = "/root/.bashrc" // bash才生效
	// 构建 docker exec 命令 // docker exec FOPS-Build sh -c "xxx.sh"
	args := []string{"exec"}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, containerId, "sh", "-c", execCmd)
	return exec.RunShellContext(ctx, "docker", args, nil, "", false)
}

// Cp 复制文件到容器内(使用Docker CLI客户端)
func (receiver container) Cp(containerId string, sourceFile, destFile string, ctx context.Context) exec.ShellWait {
	wait := receiver.Exec(containerId, "mkdir -p "+path.Dir(destFile), nil, ctx)
	wait.Wait()

	// docker cp /var/lib/fops/dist/Dockerfile FOPS-Build:/var/lib/fops/dist/Dockerfile
	args := []string{"cp", sourceFile, containerId + ":" + destFile}
	return exec.RunShellContext(ctx, "docker", args, nil, "", false)
}

// Logs 获取日志
func (receiver container) Logs(containerId string, tailCount int) collections.List[string] {
	// curl --unix-socket /var/run/docker.sock http://localhost/containers/9e76ea4b0231/logs?stdout=true&stderr=true&tail=100
	// 1. 构造 URL
	// tail=N : 告诉 Docker 守护进程只返回最后 N 行
	// stdout=true&stderr=true : 包含标准输出和错误
	url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=true&stderr=true&tail=%d", containerId, tailCount)

	// 2. 发送请求
	resp, err := receiver.unixClient.Get(url)
	if err != nil {
		return collections.List[string]{}
	}
	defer resp.Body.Close()

	// 3. 解析日志流
	// Docker 日志流格式：[8字节头] + [有效载荷] 循环
	var result collections.List[string]
	header := make([]byte, 8) // 8字节头缓冲区

	for {
		// 读取帧头
		_, err := io.ReadFull(resp.Body, header)
		if err == io.EOF || err != nil {
			break
		}

		// 解析帧头中的载荷长度 (后4字节，大端序)
		payloadLen := int(binary.BigEndian.Uint32(header[4:8]))
		if payloadLen == 0 {
			continue
		}

		// 读取载荷内容
		payload := make([]byte, payloadLen)
		_, err = io.ReadFull(resp.Body, payload)
		if err != nil {
			break
		}

		// 4. 处理内容
		// 注意：一个载荷可能包含多行日志（比如应用一次打印了换行符）
		// 我们需要按换行符拆分，保证 List 中每一项是一行
		lines := strings.Split(string(payload), "\n")
		for _, line := range lines {
			// 去除末尾可能存在的回车符，并忽略空行
			line = strings.TrimRight(line, "\r")
			if line != "" {
				result.Add(line)
			}
		}
	}

	return result
}

// Container 容器信息
type Container struct {
	ID      string        `json:"Id"`
	Name    string        // 容器名称 fops-agent
	Names   []string      `json:"Names"` //"/fops-agent.n1l83l080baf7fqs3frax2fwe.15rau7t52pgjly576wu23279z"
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
	Pid int // 需要通过Inspect(c.ID)手动获取
}

// List 获取容器列表
func (receiver container) List(status string, labels map[string]string) (collections.List[Container], error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/containers/json?status=
	url := "http://localhost/containers/json?status=" + status
	for k, v := range labels {
		url += "&label=" + k + "=" + v
	}

	containers, err := UnixGetDecode[collections.List[Container]](receiver.unixClient, url)
	containers.Foreach(func(item *Container) {
		if len(item.Names) > 0 {
			item.Name = strings.TrimPrefix(strings.Split(item.Names[0], ".")[0], "/")
		}

	})
	return containers, err
}

// Inspect 查看容器详情
func (receiver container) Inspect(containerId string) (ContainerIdInspectJson, error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/containers/kb44fvovlg1o/json
	url := fmt.Sprintf("http://localhost/containers/%s/json", containerId)

	// 1. 使用工具函数直接请求并解析
	// 注意：您的工具函数忽略了 HTTP 错误码和 JSON 解析错误，这里直接使用
	result, _ := UnixGetDecode[ContainerIdInspectJson](receiver.unixClient, url)

	// 2. 通过判断 ID 是否为空来确定容器是否存在
	// Docker 404 错误返回的是 {"message": "..."}，解析后 ID 字段为空
	if result.ID == "" {
		return result, nil
	}

	return result, nil
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
	stats, err := UnixGetDecode[StatsResponse](receiver.unixClient, url)
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

// GetFileSize 获取容器内文件大小
func (receiver container) GetFileSize(containerID, filePath string, ctx context.Context) (int64, error) {
	cmd := fmt.Sprintf("stat -c '%%s' %s", filePath)
	wait := receiver.Exec(containerID, cmd, nil, ctx)
	output, _ := wait.WaitToList()

	if output.Count() > 0 {
		return parse.ToInt64(strings.TrimSpace(output.First())), nil
	}

	return 0, nil
}

// ReadFileFromContainer 使用 docker archive API 从容器读取文件
func (receiver container) ReadFileFromContainer(containerID, filePath string, ctx context.Context) ([]byte, error) {
	resp, err := receiver.unixClient.Get(fmt.Sprintf("http://localhost/containers/%s/archive?path=%s", containerID, filePath))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("读取文件失败: %s - %s", resp.Status, string(errBody))
	}

	// 解析 tar 归档
	tarReader := tar.NewReader(resp.Body)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取tar归档失败: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tarReader); err != nil {
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("tar归档中未找到文件")
}

// ReadFileFromContainerByOffset 从容器读取文件内容（从指定行数开始）
func (receiver container) ReadFileFromContainerByOffset(containerID, filePath string, offset int64, ctx context.Context) collections.List[string] {
	// 使用 tail 命令从指定位置读取
	// tail -c +N 表示从第 N 行开始读取
	cmd := fmt.Sprintf("tail -n +%d %s 2>/dev/null", offset+1, filePath)
	// docker exec 990870bf457b sh -c "tail -n +1 /var/log/flog/fops/2026-03-07-21_1.log 2>/dev/null"
	wait := receiver.Exec(containerID, cmd, nil, ctx)
	lines, _ := wait.WaitToList()
	return lines
}

// DeleteFile 删除容器内的文件
func (receiver container) DeleteFile(containerID, filePath string, ctx context.Context) {
	cmd := fmt.Sprintf("rm -f %s", filePath)
	wait := receiver.Exec(containerID, cmd, nil, ctx)
	wait.Wait()
}

// FileExists 检查容器内文件是否存在
func (receiver container) FileExists(containerID, filePath string, ctx context.Context) bool {
	cmd := fmt.Sprintf("test -f %s && echo 'exists' || echo 'not_exists'", filePath)
	wait := receiver.Exec(containerID, cmd, nil, ctx)
	output, _ := wait.WaitToList()

	return output.Count() > 0 && strings.TrimSpace(output.First()) == "exists"
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string    // 文件名称
	Path    string    // 文件地址
	Size    int64     // 文件大小
	ModTime time.Time //
}

// ListLogFiles 列出容器内的日志文件
func (receiver container) ListLogFiles(containerID, dirPath string, fileExtension string, limitFileCount int, ctx context.Context) (collections.List[FileInfo], error) {
	// 使用 find + xargs stat 批量获取信息，兼容 Alpine/BusyBox
	var cmd string
	if limitFileCount > 0 {
		cmd = fmt.Sprintf("find %s -name '*.%s' -type f 2>/dev/null | head -n %d | xargs stat -c '%%Y %%s %%n' 2>/dev/null", dirPath, fileExtension, limitFileCount)
	} else {
		cmd = fmt.Sprintf("find %s -name '*.%s' -type f 2>/dev/null | xargs stat -c '%%Y %%s %%n' 2>/dev/null", dirPath, fileExtension)
	}

	wait := receiver.Exec(containerID, cmd, nil, ctx)
	lines, _ := wait.WaitToList()

	files := collections.NewList[FileInfo]()
	lines.Foreach(func(item *string) {
		line := strings.TrimSpace(*item)
		if line == "" {
			return
		}

		// 解析格式: 修改时间 大小 文件路径
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			return
		}

		var modTimeInt, size int64
		fmt.Sscanf(parts[0], "%d", &modTimeInt)
		fmt.Sscanf(parts[1], "%d", &size)
		filePath := strings.TrimSpace(parts[2])

		if filePath == "" {
			return
		}

		files.Add(FileInfo{
			Name:    filepath.Base(filePath),
			Path:    filePath,
			ModTime: time.Unix(modTimeInt, 0),
			Size:    size,
		})
	})

	return files, nil
}
