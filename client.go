package docker

import (
	"regexp"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/exec"
)

// Client docker client
type Client struct {
	//dockerClient *client.Client
	Container container
	Service   service
	Node      node
	Hub       hub
	Images    images
	Event     event
}

// NewClient 实例化一个Client
func NewClient() *Client {
	client := &Client{}
	client.SetChar(make(chan string, 10000))
	return client
}

// 设置接收消息的通道
func (receiver *Client) SetChar(c chan string) {
	receiver.Container.progress = c
	receiver.Service.progress = c
	receiver.Node.progress = c
	receiver.Hub.progress = c
	receiver.Images.progress = c
	receiver.Event.progress = c
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
		// 7da109011988|0.00%|7.906MiB / 3.881GiB|0.20%
		sers := strings.Split(*service, "|")
		if len(sers) != 4 {
			return
		}
		dockerStatsVO := DockerStatsVO{
			ContainerID:        sers[0],
			CpuUsagePercent:    parse.ToFloat64(strings.ReplaceAll(sers[1], "%", "")),
			MemoryUsagePercent: parse.ToFloat64(strings.ReplaceAll(sers[3], "%", "")),
		}

		// 33.36MiB / 7.586GiB
		memorys := strings.Split(sers[2], " / ")
		if len(memorys) == 2 {
			// 内存已使用（MB）memorys[0]
			if strings.Contains(memorys[0], "MiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "MiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(parse.ToFloat64(memorys[0]))
			} else if strings.Contains(memorys[0], "GiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "GiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(parse.ToFloat64(memorys[0])) * 1024
			} else if strings.Contains(memorys[0], "KiB") {
				memorys[0] = strings.ReplaceAll(memorys[0], "KiB", "")
				dockerStatsVO.MemoryUsage = parse.ToUInt64(parse.ToFloat64(memorys[0])) / 1024
			}

			// 内存限制（MB）memorys[1]
			if strings.Contains(memorys[1], "MiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "MiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(parse.ToFloat64(memorys[1]))
			} else if strings.Contains(memorys[1], "GiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "GiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(parse.ToFloat64(memorys[1]) * 1024)
			} else if strings.Contains(memorys[1], "KiB") {
				memorys[1] = strings.ReplaceAll(memorys[1], "KiB", "")
				dockerStatsVO.MemoryLimit = parse.ToUInt64(parse.ToFloat64(memorys[1]) / 1024)
			}
		}
		lstDockerInstance.Add(dockerStatsVO)
	})
	return lstDockerInstance
}

// 当前节点是否为主节点
func (receiver Client) IsMaster() bool {
	receiveOutput := make(chan string, 100)
	exec.RunShell("docker info --format '{{.Swarm.ControlAvailable}}'", receiveOutput, nil, "", false)
	lst := collections.NewListFromChan(receiveOutput)
	return lst.Contains("true")
}

// 获取主机IP
func (receiver Client) GetHostIP() string {
	receiveOutput := make(chan string, 100)
	exec.RunShell("docker info --format '{{.Swarm.NodeAddr}}'", receiveOutput, nil, "", false)
	lst := collections.NewListFromChan(receiveOutput)
	return lst.First()
}

// 获取主机IP
func (receiver Client) GetHostName() string {
	receiveOutput := make(chan string, 100)
	exec.RunShell("docker info --format '{{.Name}}'", receiveOutput, nil, "", false)
	lst := collections.NewListFromChan(receiveOutput)
	return lst.First()
}

// 获取主机信息
func (receiver Client) GetInfo() DockerInfo {
	receiveOutput := make(chan string, 100)
	exec.RunShell("docker info --format '{\"NodeAddr\":\"{{.Swarm.NodeAddr}}\",\"HostName\":\"{{.Name}}\",\"IsMaster\":{{.Swarm.ControlAvailable}},\"Version\":\"{{.ServerVersion}}\"}'", receiveOutput, nil, "", false)
	json := collections.NewListFromChan(receiveOutput).First()
	var dockerInfo DockerInfo
	snc.Unmarshal([]byte(json), &dockerInfo)
	return dockerInfo
}

type DockerInfo struct {
	NodeAddr string `json:"NodeAddr"`
	HostName string `json:"HostName"`
	IsMaster bool   `json:"IsMaster"`
	Version  string `json:"Version"`
}
