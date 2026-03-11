package docker

import (
	"net"
	"net/http"
	"time"

	"github.com/farseer-go/collections"
)

// Client docker client
type Client struct {
	//dockerClient *client.Client
	Container  container
	Service    service
	Node       node
	Hub        hub
	Images     images
	Event      event
	unixClient *http.Client
	Task       task
	Config     config
}

// NewClient 实例化一个Client
func NewClient() *Client {
	unixClient := &http.Client{
		Transport: &http.Transport{
			// 自定义 Dial 函数，将 HTTP 请求通过 Unix Socket 发送
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	client := &Client{
		unixClient: unixClient,
		Container:  container{unixClient: unixClient},
		Service:    service{unixClient: unixClient},
		Node:       node{unixClient: unixClient},
		Hub:        hub{unixClient: unixClient},
		Images:     images{unixClient: unixClient},
		Event:      event{unixClient: unixClient},
		Task:       task{unixClient: unixClient},
		Config:     config{unixClient: unixClient},
	}
	return client
}

// GetVersion 获取系统Docker版本
func (receiver *Client) GetVersion() string {
	// curl --unix-socket /var/run/docker.sock http://localhost/version
	type VersionResponse struct {
		Version    string `json:"Version"`    // Docker 版本
		ApiVersion string `json:"ApiVersion"` // API 版本
	}
	version, _ := UnixGetDecode[VersionResponse](receiver.unixClient, "http://localhost/version")
	return version.Version
}

// Stats 获取所有容器的资源使用
func (receiver *Client) Stats(containers collections.List[string]) collections.List[DockerStatsVO] {
	lstDockerInstance := collections.NewList[DockerStatsVO]()
	// 获取所有容器列表
	containers.Parallel(10, func(cID *string) {
		dockerStatsVO := receiver.Container.Stats(*cID)
		lstDockerInstance.Add(dockerStatsVO)
	})
	return lstDockerInstance
}

type DockerInfo struct {
	Name              string    `json:"Name"`              // 主机名称
	Containers        int       `json:"Containers"`        // 当前运行的容器数量
	ContainersRunning int       `json:"ContainersRunning"` // 当前正在运行的容器数量
	ContainersPaused  int       `json:"ContainersPaused"`  // 当前暂停的容器数量
	ContainersStopped int       `json:"ContainersStopped"` // 当前停止的容器数量
	Images            int       `json:"Images"`            // 当前系统中镜像的数量
	Driver            string    `json:"Driver"`            // 存储驱动 overlay2
	SystemTime        time.Time `json:"SystemTime"`        // 系统时间
	LoggingDriver     string    `json:"LoggingDriver"`     // 日志驱动 json-file
	CgroupDriver      string    `json:"CgroupDriver"`      // cgroup驱动 cgroupfs
	CgroupVersion     string    `json:"CgroupVersion"`     // cgroup版本 2
	NEventsListener   int       `json:"NEventsListener"`   // 当前监听docker事件的数量
	KernelVersion     string    `json:"KernelVersion"`     // 内核版本 5.15.0-76-generic
	OperatingSystem   string    `json:"OperatingSystem"`   // 操作系统 Ubuntu 22.04.3 LTS
	OSVersion         string    `json:"OSVersion"`         // 操作系统版本 2024.04
	OSType            string    `json:"OSType"`            // 操作系统类型 linux
	Architecture      string    `json:"Architecture"`      // 系统架构 x86_64
	NCPU              int       `json:"NCPU"`              // CPU数量 8
	MemTotal          int64     `json:"MemTotal"`          // 内存总量，单位字节
	ServerVersion     string    `json:"ServerVersion"`     // Docker版本
	Swarm             struct {
		NodeID           string     `json:"NodeID"`           // 当前节点ID
		NodeAddr         string     `json:"NodeAddr"`         // 当前节点IP地址
		ControlAvailable bool       `json:"ControlAvailable"` // 对应 IsMaster
		LocalNodeState   string     `json:"LocalNodeState"`   // active - 当前节点状态
		RemoteManagers   []struct { // 集群中远程管理节点的信息
			NodeID string `json:"NodeID"` // 远程管理节点的ID
			Addr   string `json:"Addr"`   // 远程管理节点的地址
		} `json:"RemoteManagers"`
		Nodes    int    `json:"Nodes"`    // 集群中节点总数
		Managers int    `json:"Managers"` // 集群中管理节点数量
		Error    string `json:"Error"`    // 集群错误信息，如果没有错误则为空字符串
	} `json:"Swarm"`
	Labels map[string]string `json:"Labels"` // Docker节点标签
}

func (receiver *Client) GetInfo() DockerInfo {
	// curl --unix-socket /var/run/docker.sock http://localhost/info
	apiData, _ := UnixGetDecode[DockerInfo](receiver.unixClient, "http://localhost/info")
	return apiData
}

// SyncConfig 检查并更新服务的配置版本
// 返回值：是否更新了配置
func (receiver *Client) SyncConfig(appName string, targetConfigPath string) bool {
	appVer, err := receiver.Service.GetCurConfigVersion(appName)
	if err != nil {
		return false
	}
	// 说明服务没有使用配置
	if appVer == 0 {
		return false
	}

	// 获取服务当前使用的配置
	configVersion, err := receiver.Config.GetLastVersion(appName)
	if err != nil {
		return false
	}

	// 没有读取到配置文件,则退出
	if configVersion.Version == 0 {
		return false
	}

	// 如果版本不一致，更新配置
	if configVersion.Version != appVer {
		isUpdate, _ := receiver.Service.UpdateServiceConfig(appName, configVersion.ID, configVersion.Spec.Name, targetConfigPath)
		return isUpdate
	}

	return false
}
