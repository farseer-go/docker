package docker

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/utils/exec"
	"regexp"
	"strings"
)

// Client docker client
type Client struct {
	//dockerClient *client.Client
	Container container
	Service   service
	Node      node
	Hub       hub
	Images    images
}

// NewClient 实例化一个Client
func NewClient() *Client {
	return &Client{}
}

// GetVersion 获取系统Docker版本
func (receiver Client) GetVersion() string {
	receiveOutput := make(chan string, 100)
	exec.RunShell("docker version --format '{{.Server.Version}}'", receiveOutput, nil, "", false)
	lst := collections.NewListFromChan(receiveOutput)
	re := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	for _, s := range lst.ToArray() {
		if re.MatchString(s) {
			return s
		}
	}
	return ""
}

// Stats 获取所有容器的资源使用
func (receiver Client) Stats() collections.List[DockerStatsVO] {
	progress := make(chan string, 1000)
	// docker stats --format "table {{.Container}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}" --no-stream
	var exitCode = exec.RunShell("docker stats --format \"table {{.Container}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}\" --no-stream", progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
	lstDockerInstance := collections.NewList[DockerStatsVO]()
	if exitCode != 0 || serviceList.Count() == 0 {
		return lstDockerInstance
	}

	// 移除标题
	serviceList.RemoveAt(0)
	serviceList.Foreach(func(service *string) {
		// 95351953ac34|0.84%|0.43%|33.36MiB / 7.586GiB
		sers := strings.Split(*service, "|")
		if len(sers) != 4 {
			return
		}
		dockerStatsVO := DockerStatsVO{
			ContainerID:        sers[0],
			CpuUsagePercent:    parse.ToFloat64(strings.ReplaceAll(sers[1], "%", "")),
			MemoryUsagePercent: parse.ToFloat64(strings.ReplaceAll(sers[2], "%", "")),
		}

		// 33.36MiB / 7.586GiB
		memorys := strings.Split(sers[3], " / ")
		if len(memorys) == 2 {
			// 内存已使用（MB）memorys[0]
			if strings.Contains(memorys[0], "MiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "MiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(memorys[0])
			} else if strings.Contains(memorys[0], "GiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "GiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(memorys[0]) * 1024
			} else if strings.Contains(memorys[0], "KiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "KiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(memorys[0]) / 1024
			}

			// 内存限制（MB）memorys[1]
			if strings.Contains(memorys[1], "MiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "MiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(memorys[1])
			} else if strings.Contains(memorys[1], "GiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "GiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(memorys[1]) * 1024
			} else if strings.Contains(memorys[1], "KiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "KiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(memorys[1]) / 1024
			}
		}
		lstDockerInstance.Add(dockerStatsVO)
	})
	return lstDockerInstance
}
