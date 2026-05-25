package docker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/collections"
)

const defaultDockerHost = "unix:///var/run/docker.sock"

var DefaultClient *Client = NewClient()

type dockerEndpoint struct {
	rawHost string
	scheme  string
	address string
	baseURL string
}

type dockerAPI struct {
	httpClient *http.Client
	endpoint   dockerEndpoint
	initErr    error
}

func (receiver *dockerAPI) URL(apiPath string) string {
	return strings.TrimRight(receiver.endpoint.baseURL, "/") + "/" + strings.TrimLeft(apiPath, "/")
}

func (receiver *dockerAPI) cliEnv(extra map[string]string) map[string]string {
	env := map[string]string{}
	for k, v := range extra {
		env[k] = v
	}
	env["DOCKER_HOST"] = receiver.endpoint.rawHost
	return env
}

type errorTransport struct {
	err error
}

func (receiver errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, receiver.err
}

// Client docker client
type Client struct {
	//dockerClient *client.Client
	Container container
	Service   service
	Node      node
	Hub       hub
	Images    images
	Event     event
	api       *dockerAPI
	Task      task
	Config    config
}

// NewClient 实例化一个Client
func NewClient() *Client {
	api := newDockerAPI(os.Getenv("DOCKER_HOST"))
	client := &Client{
		api:       api,
		Container: container{api: api},
		Service:   service{api: api},
		Node:      node{api: api},
		Hub:       hub{api: api},
		Images:    images{api: api},
		Event:     event{api: api},
		Task:      task{api: api},
		Config:    config{api: api},
	}
	return client
}

func newDockerAPI(rawHost string) *dockerAPI {
	endpoint, err := parseDockerHost(rawHost)
	if err != nil {
		return &dockerAPI{
			httpClient: &http.Client{Transport: errorTransport{err: err}},
			endpoint:   endpoint,
			initErr:    err,
		}
	}

	return &dockerAPI{
		httpClient: newHTTPClient(endpoint),
		endpoint:   endpoint,
	}
}

func parseDockerHost(rawHost string) (dockerEndpoint, error) {
	rawHost = strings.TrimSpace(rawHost)
	if rawHost == "" {
		rawHost = defaultDockerHost
	}

	endpoint := dockerEndpoint{rawHost: rawHost}
	if dockerTLSVerifyEnabled() {
		return endpoint, fmt.Errorf("unsupported Docker TLS configuration for DOCKER_HOST %q", rawHost)
	}

	u, err := url.Parse(rawHost)
	if err != nil {
		return endpoint, err
	}

	switch u.Scheme {
	case "unix":
		if u.Path == "" {
			return endpoint, fmt.Errorf("invalid DOCKER_HOST %q: unix socket path is empty", rawHost)
		}
		endpoint.scheme = "unix"
		endpoint.address = u.Path
		endpoint.baseURL = "http://docker"
		return endpoint, nil
	case "tcp":
		if u.Host == "" {
			return endpoint, fmt.Errorf("invalid DOCKER_HOST %q: tcp host is empty", rawHost)
		}
		endpoint.scheme = "tcp"
		endpoint.address = u.Host
		endpoint.baseURL = "http://" + u.Host
		return endpoint, nil
	case "http":
		if u.Host == "" {
			return endpoint, fmt.Errorf("invalid DOCKER_HOST %q: http host is empty", rawHost)
		}
		endpoint.scheme = "tcp"
		endpoint.address = u.Host
		endpoint.baseURL = "http://" + u.Host
		return endpoint, nil
	case "https":
		return endpoint, fmt.Errorf("unsupported DOCKER_HOST scheme %q; supported schemes: unix, tcp, http", u.Scheme)
	case "":
		return endpoint, fmt.Errorf("invalid DOCKER_HOST %q: scheme is empty", rawHost)
	default:
		return endpoint, fmt.Errorf("unsupported DOCKER_HOST scheme %q; supported schemes: unix, tcp, http", u.Scheme)
	}
}

func dockerTLSVerifyEnabled() bool {
	value := strings.TrimSpace(os.Getenv("DOCKER_TLS_VERIFY"))
	return value != "" && value != "0"
}

func newHTTPClient(endpoint dockerEndpoint) *http.Client {
	transport := &http.Transport{IdleConnTimeout: 90 * time.Second}
	if endpoint.scheme == "unix" {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", endpoint.address)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{Transport: transport}
}

// GetVersion 获取系统Docker版本
func (receiver *Client) GetVersion() string {
	type VersionResponse struct {
		Version    string `json:"Version"`    // Docker 版本
		ApiVersion string `json:"ApiVersion"` // API 版本
	}
	version, _ := UnixGetDecode[VersionResponse](receiver.api.httpClient, receiver.api.URL("/version"))
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
	apiData, _ := UnixGetDecode[DockerInfo](receiver.api.httpClient, receiver.api.URL("/info"))
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

// 是否运行在docker容器内
func (receiver *Client) IsRunInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
