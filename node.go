package docker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/farseer-go/collections"
)

type node struct {
	unixClient *http.Client
}

// DockerNodeVO 集群节点信息 docker node ls
type DockerNodeVO struct {
	ID          string `json:"ID"` // 节点ID
	Description struct {
		Hostname string `json:"Hostname"` // 节点名称
		Platform struct {
			Architecture string `json:"Architecture"`
			OS           string `json:"OS"`
		} `json:"Platform"`
		Resources struct {
			NanoCPUs           int64    `json:"NanoCPUs"`
			MemoryBytes        int64    `json:"MemoryBytes"` // 内存总容量（Bytes）
			Memory             string   // 内存总容量（GB）
			MemoryUsagePercent float64  // 内存使用百分比
			MemoryUsage        float64  // 内存已使用（MB）
			CpuUsagePercent    float64  // CPU使用百分比(本地计算)
			DiskTotal          string   // 硬盘总容量（GB）
			Disk               []DiskVO // 硬盘容量
		} `json:"Resources"`
	} `json:"Description"`
	Status struct {
		State string `json:"State"` // 主机状态   ready
		Addr  string `json:"Addr"`  // 主机IP
	} `json:"Status"`
	Spec struct {
		Labels       map[string]string `json:"Labels"`       // 标签
		Role         string            `json:"Role"`         // 节点角色   manager worker
		Availability string            `json:"Availability"` // 节点状态   active pause drain
	} `json:"Spec"`
	ManagerStatus struct {
		Leader       bool   `json:"Leader"`       // 是否是管理节点
		Reachability string `json:"Reachability"` // 管理节点状态   reachable unreachable
		Addr         string `json:"Addr"`         // 管理节点IP	 192.168.1.2:2377
	} `json:"ManagerStatus"`
	Engine struct {
		EngineVersion string `json:"EngineVersion"` // 引擎版本
	} `json:"Engine"`
	CreatedAt time.Time `json:"CreatedAt"` // 创建时间
	UpdatedAt time.Time `json:"UpdatedAt"` // 更新时间

	IsHealth bool                            // 应用是否健康
	Label    collections.List[DockerLabelVO] // 标签
}

// DockerLabelVO 标签
type DockerLabelVO struct {
	Name  string // 标签名称
	Value string // 标签值
}

// List 获取主机节点列表
func (receiver node) List() collections.List[DockerNodeVO] {
	// curl --unix-socket /var/run/docker.sock http://localhost/nodes
	nodesUrl := "http://localhost/nodes"
	nodes, err := UnixGetDecode[collections.List[DockerNodeVO]](receiver.unixClient, nodesUrl)
	if err != nil {
		return nodes
	}

	// 更新健康状态
	nodes.Foreach(func(item *DockerNodeVO) {
		item.IsHealth = item.Status.State == "Ready" && item.Spec.Availability == "Active"
		item.Description.Resources.NanoCPUs = item.Description.Resources.NanoCPUs / 1000000000
		item.Description.Resources.Memory = fmt.Sprintf("%.2fGB", float64(item.Description.Resources.MemoryBytes)/1024/1024/1024)

		// 将标签转换为列表
		item.Label = collections.NewList[DockerLabelVO]()
		for k, v := range item.Spec.Labels {
			item.Label.Add(DockerLabelVO{
				Name:  k,
				Value: v,
			})
		}
	})

	return nodes
}

// Info 获取节点详情
func (receiver node) Info(nodeName string) DockerNodeVO {
	// 2. 调用接口
	// 直接请求 /nodes/{name}，返回单个对象
	url := fmt.Sprintf("http://localhost/nodes/%s", nodeName)

	// 使用泛型获取数据
	nodeData, _ := UnixGetDecode[DockerNodeVO](receiver.unixClient, url)

	// 3. 转换为业务 VO
	nodeData.Label = collections.NewList[DockerLabelVO]()

	// 如果关键数据为空，说明节点不存在
	if nodeData.Status.Addr == "" {
		return nodeData
	}

	// 处理 CPU (API 返回 NanoCPUs，需要除以 1e9)
	if nodeData.Description.Resources.NanoCPUs > 0 {
		nodeData.Description.Resources.NanoCPUs = nodeData.Description.Resources.NanoCPUs / 1000000000
	}

	// 处理 Memory (API 返回 Bytes，这里转为 GB，保留2位小数)
	if nodeData.Description.Resources.MemoryBytes > 0 {
		nodeData.Description.Resources.Memory = fmt.Sprintf("%.2fGB", float64(nodeData.Description.Resources.MemoryBytes)/1024/1024/1024)
	}

	// 处理 Labels
	// API 返回的是 map[string]string，转换为 List[DockerLabelVO]
	for k, v := range nodeData.Spec.Labels {
		nodeData.Label.Add(DockerLabelVO{
			Name:  k,
			Value: v,
		})
	}

	return nodeData
}
