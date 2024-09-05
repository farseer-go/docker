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
	NodeName           string                          // 节点名称
	Status             string                          // 主机状态   Ready
	Availability       string                          // 节点状态
	IsMaster           bool                            // 是否为主节点
	IsHealth           bool                            // 应用是否健康
	EngineVersion      string                          // 引擎版本
	IP                 string                          // 节点IP
	OS                 string                          // 操作系统
	Architecture       string                          // 架构
	CPUs               string                          // CPU核心数
	Memory             string                          // 内存
	Label              collections.List[DockerLabelVO] // 标签
	AgentIP            string                          // 代理容器IP
	CpuUsagePercent    float64                         // CPU使用百分比
	MemoryUsagePercent float64                         // 内存使用百分比
	MemoryUsage        float64                         // 内存已使用（MB）
	Disk               uint64                          // 硬盘总容量（GB）
	DiskUsagePercent   float64                         // 硬盘使用百分比
	DiskUsage          float64                         // 硬盘已用空间（GB）
}

// DockerLabelVO 标签
type DockerLabelVO struct {
	Name  string // 标签名称
	Value string // 标签值
}

// DockerStatsVO 容器的状态
type DockerStatsVO struct {
	ContainerID        string  // 容器ID
	CpuUsagePercent    float64 // CPU使用百分比
	MemoryUsagePercent float64 // 内存使用百分比
	MemoryUsage        uint64  // 内存已使用（MB）
	MemoryLimit        uint64  // 内存限制（MB）
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

type ServiceIdInspectJson []struct {
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
		Err             string    `json:"Err"`
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

type ContainerIdInspectJson []struct {
	ID      string        `json:"Id"`
	Created time.Time     `json:"Created"`
	Path    string        `json:"Path"`
	Args    []interface{} `json:"Args"`
	State   struct {
		Status     string    `json:"Status"`
		Running    bool      `json:"Running"`
		Paused     bool      `json:"Paused"`
		Restarting bool      `json:"Restarting"`
		OOMKilled  bool      `json:"OOMKilled"`
		Dead       bool      `json:"Dead"`
		Pid        int       `json:"Pid"`
		ExitCode   int       `json:"ExitCode"`
		Error      string    `json:"Error"`
		StartedAt  time.Time `json:"StartedAt"`
		FinishedAt time.Time `json:"FinishedAt"`
	} `json:"State"`
	Image           string      `json:"Image"`
	ResolvConfPath  string      `json:"ResolvConfPath"`
	HostnamePath    string      `json:"HostnamePath"`
	HostsPath       string      `json:"HostsPath"`
	LogPath         string      `json:"LogPath"`
	Name            string      `json:"Name"`
	RestartCount    int         `json:"RestartCount"`
	Driver          string      `json:"Driver"`
	Platform        string      `json:"Platform"`
	MountLabel      string      `json:"MountLabel"`
	ProcessLabel    string      `json:"ProcessLabel"`
	AppArmorProfile string      `json:"AppArmorProfile"`
	ExecIDs         interface{} `json:"ExecIDs"`
	HostConfig      struct {
		Binds           interface{} `json:"Binds"`
		ContainerIDFile string      `json:"ContainerIDFile"`
		LogConfig       struct {
			Type   string `json:"Type"`
			Config struct {
				MaxFile string `json:"max-file"`
				MaxSize string `json:"max-size"`
			} `json:"Config"`
		} `json:"LogConfig"`
		NetworkMode  string `json:"NetworkMode"`
		PortBindings struct {
		} `json:"PortBindings"`
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
		AutoRemove           bool          `json:"AutoRemove"`
		VolumeDriver         string        `json:"VolumeDriver"`
		VolumesFrom          interface{}   `json:"VolumesFrom"`
		CapAdd               interface{}   `json:"CapAdd"`
		CapDrop              interface{}   `json:"CapDrop"`
		CgroupnsMode         string        `json:"CgroupnsMode"`
		DNS                  interface{}   `json:"Dns"`
		DNSOptions           interface{}   `json:"DnsOptions"`
		DNSSearch            interface{}   `json:"DnsSearch"`
		ExtraHosts           interface{}   `json:"ExtraHosts"`
		GroupAdd             interface{}   `json:"GroupAdd"`
		IpcMode              string        `json:"IpcMode"`
		Cgroup               string        `json:"Cgroup"`
		Links                interface{}   `json:"Links"`
		OomScoreAdj          int           `json:"OomScoreAdj"`
		PidMode              string        `json:"PidMode"`
		Privileged           bool          `json:"Privileged"`
		PublishAllPorts      bool          `json:"PublishAllPorts"`
		ReadonlyRootfs       bool          `json:"ReadonlyRootfs"`
		SecurityOpt          interface{}   `json:"SecurityOpt"`
		UTSMode              string        `json:"UTSMode"`
		UsernsMode           string        `json:"UsernsMode"`
		ShmSize              int           `json:"ShmSize"`
		Runtime              string        `json:"Runtime"`
		ConsoleSize          []int         `json:"ConsoleSize"`
		Isolation            string        `json:"Isolation"`
		CPUShares            int           `json:"CpuShares"`
		Memory               int           `json:"Memory"`
		NanoCpus             int           `json:"NanoCpus"`
		CgroupParent         string        `json:"CgroupParent"`
		BlkioWeight          int           `json:"BlkioWeight"`
		BlkioWeightDevice    interface{}   `json:"BlkioWeightDevice"`
		BlkioDeviceReadBps   interface{}   `json:"BlkioDeviceReadBps"`
		BlkioDeviceWriteBps  interface{}   `json:"BlkioDeviceWriteBps"`
		BlkioDeviceReadIOps  interface{}   `json:"BlkioDeviceReadIOps"`
		BlkioDeviceWriteIOps interface{}   `json:"BlkioDeviceWriteIOps"`
		CPUPeriod            int           `json:"CpuPeriod"`
		CPUQuota             int           `json:"CpuQuota"`
		CPURealtimePeriod    int           `json:"CpuRealtimePeriod"`
		CPURealtimeRuntime   int           `json:"CpuRealtimeRuntime"`
		CpusetCpus           string        `json:"CpusetCpus"`
		CpusetMems           string        `json:"CpusetMems"`
		Devices              interface{}   `json:"Devices"`
		DeviceCgroupRules    interface{}   `json:"DeviceCgroupRules"`
		DeviceRequests       interface{}   `json:"DeviceRequests"`
		KernelMemory         int           `json:"KernelMemory"`
		KernelMemoryTCP      int           `json:"KernelMemoryTCP"`
		MemoryReservation    int           `json:"MemoryReservation"`
		MemorySwap           int           `json:"MemorySwap"`
		MemorySwappiness     interface{}   `json:"MemorySwappiness"`
		OomKillDisable       bool          `json:"OomKillDisable"`
		PidsLimit            interface{}   `json:"PidsLimit"`
		Ulimits              []interface{} `json:"Ulimits"`
		CPUCount             int           `json:"CpuCount"`
		CPUPercent           int           `json:"CpuPercent"`
		IOMaximumIOps        int           `json:"IOMaximumIOps"`
		IOMaximumBandwidth   int           `json:"IOMaximumBandwidth"`
		Mounts               []struct {
			Type   string `json:"Type"`
			Source string `json:"Source"`
			Target string `json:"Target"`
		} `json:"Mounts"`
		MaskedPaths   []string `json:"MaskedPaths"`
		ReadonlyPaths []string `json:"ReadonlyPaths"`
		Init          bool     `json:"Init"`
	} `json:"HostConfig"`
	GraphDriver struct {
		Data struct {
			LowerDir  string `json:"LowerDir"`
			MergedDir string `json:"MergedDir"`
			UpperDir  string `json:"UpperDir"`
			WorkDir   string `json:"WorkDir"`
		} `json:"Data"`
		Name string `json:"Name"`
	} `json:"GraphDriver"`
	Mounts []struct {
		Type        string `json:"Type"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
		Propagation string `json:"Propagation"`
	} `json:"Mounts"`
	Config struct {
		Hostname     string            `json:"Hostname"`
		Domainname   string            `json:"Domainname"`
		User         string            `json:"User"`
		AttachStdin  bool              `json:"AttachStdin"`
		AttachStdout bool              `json:"AttachStdout"`
		AttachStderr bool              `json:"AttachStderr"`
		Tty          bool              `json:"Tty"`
		OpenStdin    bool              `json:"OpenStdin"`
		StdinOnce    bool              `json:"StdinOnce"`
		Env          []string          `json:"Env"`
		Cmd          interface{}       `json:"Cmd"`
		Image        string            `json:"Image"`
		Volumes      interface{}       `json:"Volumes"`
		WorkingDir   string            `json:"WorkingDir"`
		Entrypoint   []string          `json:"Entrypoint"`
		OnBuild      interface{}       `json:"OnBuild"`
		Labels       map[string]string `json:"Labels"`
	} `json:"Config"`
	NetworkSettings struct {
		Bridge                 string `json:"Bridge"`
		SandboxID              string `json:"SandboxID"`
		HairpinMode            bool   `json:"HairpinMode"`
		LinkLocalIPv6Address   string `json:"LinkLocalIPv6Address"`
		LinkLocalIPv6PrefixLen int    `json:"LinkLocalIPv6PrefixLen"`
		Ports                  struct {
		} `json:"Ports"`
		SandboxKey             string      `json:"SandboxKey"`
		SecondaryIPAddresses   interface{} `json:"SecondaryIPAddresses"`
		SecondaryIPv6Addresses interface{} `json:"SecondaryIPv6Addresses"`
		EndpointID             string      `json:"EndpointID"`
		Gateway                string      `json:"Gateway"`
		GlobalIPv6Address      string      `json:"GlobalIPv6Address"`
		GlobalIPv6PrefixLen    int         `json:"GlobalIPv6PrefixLen"`
		IPAddress              string      `json:"IPAddress"`
		IPPrefixLen            int         `json:"IPPrefixLen"`
		IPv6Gateway            string      `json:"IPv6Gateway"`
		MacAddress             string      `json:"MacAddress"`
		Networks               struct {
			Net struct {
				IPAMConfig struct {
					IPv4Address string `json:"IPv4Address"`
				} `json:"IPAMConfig"`
				Links               interface{} `json:"Links"`
				Aliases             []string    `json:"Aliases"`
				NetworkID           string      `json:"NetworkID"`
				EndpointID          string      `json:"EndpointID"`
				Gateway             string      `json:"Gateway"`
				IPAddress           string      `json:"IPAddress"`
				IPPrefixLen         int         `json:"IPPrefixLen"`
				IPv6Gateway         string      `json:"IPv6Gateway"`
				GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
				GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
				MacAddress          string      `json:"MacAddress"`
				DriverOpts          interface{} `json:"DriverOpts"`
			} `json:"net"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}
