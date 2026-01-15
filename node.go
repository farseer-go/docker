package docker

import (
	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
)

type node struct {
	//progress chan string
}

// List 获取主机节点列表
func (receiver node) List() collections.List[DockerNodeVO] {
	// docker node ls --format "table {{.Hostname}}|{{.Status}}|{{.Availability}}|{{.ManagerStatus}}|{{.EngineVersion}}"
	serviceList, exitCode := exec.RunShellCommand("docker node ls --format \"table {{.Hostname}}|{{.Status}}|{{.Availability}}|{{.ManagerStatus}}|{{.EngineVersion}}\"", nil, "", false)
	lstDockerInstance := collections.NewList[DockerNodeVO]()
	if exitCode != 0 || serviceList.Count() == 0 {
		return lstDockerInstance
	}

	// 移除标题
	serviceList.RemoveAt(0)
	serviceList.Foreach(func(service *string) {
		// test|Ready|Active|Leader|20.10.17
		sers := strings.Split(*service, "|")
		if len(sers) < 5 {
			return
		}
		lstDockerInstance.Add(DockerNodeVO{
			NodeName:      sers[0],
			Status:        sers[1],
			Availability:  sers[2],
			IsMaster:      sers[3] == "Leader",
			EngineVersion: sers[4],
			IsHealth:      sers[1] == "Ready" && sers[2] == "Active",
		})
	})
	return lstDockerInstance
}

// Info 获取节点详情
func (receiver node) Info(nodeName string) DockerNodeVO {
	// docker node inspect node_1 --pretty
	serviceList, exitCode := exec.RunShellCommand(fmt.Sprintf("docker node inspect %s --pretty", nodeName), nil, "", false)
	vo := DockerNodeVO{
		Label: collections.NewList[DockerLabelVO](),
	}
	if exitCode != 0 || serviceList.Count() == 0 {
		return vo
	}
	serviceList.For(func(index int, item *string) {
		kv := strings.Split(*item, ":")
		if len(kv) != 2 {
			return
		}
		name := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch name {
		case "Address":
			// 跳过Manager Status
			if strings.Contains(val, ":") {
				return
			}
			vo.IP = val
		case "Operating System":
			vo.OS = val
		case "Architecture":
			vo.Architecture = val
		case "CPUs":
			vo.CPUs = val
		case "Memory":
			vo.Memory = val
		case "Labels":
			// 标签要特殊处理
			/*
			   Labels:
			    - run=job
			    - type=master
			*/
			tag := " - "
			for {
				index++
				content := serviceList.Index(index)
				if !strings.HasPrefix(content, tag) {
					return
				}
				// 移除标签
				content = strings.TrimSpace(content[len(tag):])
				kvs := strings.Split(content, "=")
				if len(kvs) > 1 {
					vo.Label.Add(DockerLabelVO{
						Name:  kvs[0],
						Value: kvs[1],
					})
				}
			}
		}
	})
	return vo
}
