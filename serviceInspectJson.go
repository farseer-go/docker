package docker

import (
	"github.com/farseer-go/collections"
	"time"
)

// ServiceListVO 容器的名称 实例数量 副本数量 镜像（docker service ls）
type ServiceListVO struct {
	Id        string // 容器ID
	Name      string // 容器名称
	Instances int    // 实例数量
	Replicas  int    // 副本数量
	Image     string // 镜像
}

// ServicePsVO 容器的实例信息 docker service ps fops
type ServicePsVO struct {
	ServiceId string // 服务ID
	Name      string // 容器名称
	Image     string // 镜像
	Node      string // 节点
	State     string // 状态   Shutdown Running
	StateInfo string // 状态
	Error     string // 错误信息
}

// DockerNodeVO 集群节点信息 docker node ls
type DockerNodeVO struct {
	NodeName      string                          // 节点名称
	Status        string                          // 主机状态   Ready
	Availability  string                          // 节点状态
	IsMaster      bool                            // 是否为主节点
	IsHealth      bool                            // 应用是否健康
	EngineVersion string                          // 引擎版本
	IP            string                          // 节点IP
	OS            string                          // 操作系统
	Architecture  string                          // 架构
	CPUs          string                          // CPU核心数
	Memory        string                          // 内存
	Label         collections.List[DockerLabelVO] // 标签
}

// DockerLabelVO 标签
type DockerLabelVO struct {
	Name  string // 标签名称
	Value string // 标签值
}

// ServiceInspectJson 服务详情
type ServiceInspectJson []struct {
	ID      string `json:"ID"`
	Version struct {
		Index int `json:"Index"`
	} `json:"Version"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	Spec      struct {
		Name         string            `json:"Name"`
		Labels       map[string]string `json:"Labels"`
		TaskTemplate struct {
			ContainerSpec struct {
				Image  string `json:"Image"`
				Init   bool   `json:"Init"`
				Mounts []struct {
					Type   string `json:"Type"`
					Source string `json:"Source"`
					Target string `json:"Target"`
				} `json:"Mounts"`
				StopGracePeriod int64 `json:"StopGracePeriod"`
				DNSConfig       struct {
				} `json:"DNSConfig"`
				Isolation string `json:"Isolation"`
			} `json:"ContainerSpec"`
			Resources struct {
				Limits struct {
				} `json:"Limits"`
				Reservations struct {
				} `json:"Reservations"`
			} `json:"Resources"`
			RestartPolicy struct {
				Condition   string `json:"Condition"`
				Delay       int64  `json:"Delay"`
				MaxAttempts int    `json:"MaxAttempts"`
			} `json:"RestartPolicy"`
			Placement struct {
				Constraints []string `json:"Constraints"`
				Platforms   []struct {
					Architecture string `json:"Architecture"`
					OS           string `json:"OS"`
				} `json:"Platforms"`
			} `json:"Placement"`
			Networks []struct {
				Target string `json:"Target"`
			} `json:"Networks"`
			ForceUpdate int    `json:"ForceUpdate"`
			Runtime     string `json:"Runtime"`
		} `json:"TaskTemplate"`
		Mode struct {
			Replicated struct {
				Replicas int `json:"Replicas"`
			} `json:"Replicated"`
		} `json:"Mode"`
		UpdateConfig struct {
			Parallelism     int    `json:"Parallelism"`
			Delay           int64  `json:"Delay"`
			FailureAction   string `json:"FailureAction"`
			Monitor         int64  `json:"Monitor"`
			MaxFailureRatio int    `json:"MaxFailureRatio"`
			Order           string `json:"Order"`
		} `json:"UpdateConfig"`
		RollbackConfig struct {
			Parallelism     int    `json:"Parallelism"`
			FailureAction   string `json:"FailureAction"`
			Monitor         int64  `json:"Monitor"`
			MaxFailureRatio int    `json:"MaxFailureRatio"`
			Order           string `json:"Order"`
		} `json:"RollbackConfig"`
		EndpointSpec struct {
			Mode string `json:"Mode"`
		} `json:"EndpointSpec"`
	} `json:"Spec"`
	PreviousSpec struct {
		Name         string            `json:"Name"`
		Labels       map[string]string `json:"Labels"`
		TaskTemplate struct {
			ContainerSpec struct {
				Image  string `json:"Image"`
				Init   bool   `json:"Init"`
				Mounts []struct {
					Type   string `json:"Type"`
					Source string `json:"Source"`
					Target string `json:"Target"`
				} `json:"Mounts"`
				DNSConfig struct {
				} `json:"DNSConfig"`
				Isolation string `json:"Isolation"`
			} `json:"ContainerSpec"`
			Resources struct {
				Limits struct {
				} `json:"Limits"`
				Reservations struct {
				} `json:"Reservations"`
			} `json:"Resources"`
			Placement struct {
				Constraints []string `json:"Constraints"`
				Platforms   []struct {
					Architecture string `json:"Architecture"`
					OS           string `json:"OS"`
				} `json:"Platforms"`
			} `json:"Placement"`
			Networks []struct {
				Target string `json:"Target"`
			} `json:"Networks"`
			ForceUpdate int    `json:"ForceUpdate"`
			Runtime     string `json:"Runtime"`
		} `json:"TaskTemplate"`
		Mode struct {
			Replicated struct {
				Replicas int `json:"Replicas"`
			} `json:"Replicated"`
		} `json:"Mode"`
		UpdateConfig struct {
			Parallelism     int    `json:"Parallelism"`
			Delay           int64  `json:"Delay"`
			FailureAction   string `json:"FailureAction"`
			Monitor         int64  `json:"Monitor"`
			MaxFailureRatio int    `json:"MaxFailureRatio"`
			Order           string `json:"Order"`
		} `json:"UpdateConfig"`
		EndpointSpec struct {
			Mode string `json:"Mode"`
		} `json:"EndpointSpec"`
	} `json:"PreviousSpec"`
	Endpoint struct {
		Spec struct {
			Mode string `json:"Mode"`
		} `json:"Spec"`
		VirtualIPs []struct {
			NetworkID string `json:"NetworkID"`
			Addr      string `json:"Addr"`
		} `json:"VirtualIPs"`
	} `json:"Endpoint"`
	UpdateStatus struct {
		State       string    `json:"State"`
		StartedAt   time.Time `json:"StartedAt"`
		CompletedAt time.Time `json:"CompletedAt"`
		Message     string    `json:"Message"`
	} `json:"UpdateStatus"`
}

type ContainerInspectJson []struct {
	ID      string `json:"ID"`
	Version struct {
		Index int `json:"Index"`
	} `json:"Version"`
	CreatedAt time.Time         `json:"CreatedAt"`
	UpdatedAt time.Time         `json:"UpdatedAt"`
	Labels    map[string]string `json:"Labels"`
	Spec      struct {
		ContainerSpec struct {
			Image  string   `json:"Image"`
			Env    []string `json:"Env"`
			Init   bool     `json:"Init"`
			Mounts []struct {
				Type   string `json:"Type"`
				Source string `json:"Source"`
				Target string `json:"Target"`
			} `json:"Mounts"`
			DNSConfig struct {
			} `json:"DNSConfig"`
			Isolation string `json:"Isolation"`
		} `json:"ContainerSpec"`
		Resources struct {
			Limits struct {
			} `json:"Limits"`
			Reservations struct {
			} `json:"Reservations"`
		} `json:"Resources"`
		Placement struct {
			Constraints []string `json:"Constraints"`
			Platforms   []struct {
				Architecture string `json:"Architecture"`
				OS           string `json:"OS"`
			} `json:"Platforms"`
		} `json:"Placement"`
		Networks []struct {
			Target string `json:"Target"`
		} `json:"Networks"`
		ForceUpdate int `json:"ForceUpdate"`
	} `json:"Spec"`
	ServiceID string `json:"ServiceID"`
	Slot      int    `json:"Slot"`
	NodeID    string `json:"NodeID"`
	Status    struct {
		Timestamp       time.Time `json:"Timestamp"`
		State           string    `json:"State"`
		Message         string    `json:"Message"`
		Err         	string    `json:"Err"`
		ContainerStatus struct {
			ContainerID string `json:"ContainerID"`
			PID         int    `json:"PID"`
			ExitCode    int    `json:"ExitCode"`
		} `json:"ContainerStatus"`
		PortStatus struct {
		} `json:"PortStatus"`
	} `json:"Status"`
	DesiredState        string `json:"DesiredState"`
	NetworksAttachments []struct {
		Network struct {
			ID      string `json:"ID"`
			Version struct {
				Index int `json:"Index"`
			} `json:"Version"`
			CreatedAt time.Time `json:"CreatedAt"`
			UpdatedAt time.Time `json:"UpdatedAt"`
			Spec      struct {
				Name   string `json:"Name"`
				Labels struct {
				} `json:"Labels"`
				DriverConfiguration struct {
					Name string `json:"Name"`
				} `json:"DriverConfiguration"`
				Attachable  bool `json:"Attachable"`
				IPAMOptions struct {
					Driver struct {
						Name string `json:"Name"`
					} `json:"Driver"`
					Configs []struct {
						Subnet  string `json:"Subnet"`
						Gateway string `json:"Gateway"`
					} `json:"Configs"`
				} `json:"IPAMOptions"`
				Scope string `json:"Scope"`
			} `json:"Spec"`
			DriverState struct {
				Name    string `json:"Name"`
				Options struct {
					ComDockerNetworkDriverOverlayVxlanidList string `json:"com.docker.network.driver.overlay.vxlanid_list"`
				} `json:"Options"`
			} `json:"DriverState"`
			IPAMOptions struct {
				Driver struct {
					Name string `json:"Name"`
				} `json:"Driver"`
				Configs []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				} `json:"Configs"`
			} `json:"IPAMOptions"`
		} `json:"Network"`
		Addresses []string `json:"Addresses"`
	} `json:"NetworksAttachments"`
}
