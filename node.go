package docker

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
	"strings"
)

type node struct {
	dockerClient *client.Client
}

// List 获取主机节点列表
func (receiver node) List() collections.List[DockerNodeVO] {
	progress := make(chan string, 1000)
	// docker node ls --format "table {{.Hostname}}|{{.Status}}|{{.Availability}}|{{.ManagerStatus}}|{{.EngineVersion}}"
	var exitCode = exec.RunShell("docker node ls --format \"table {{.Hostname}}|{{.Status}}|{{.Availability}}|{{.ManagerStatus}}|{{.EngineVersion}}\"", progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
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
		})
	})
	return lstDockerInstance
}

// Info 获取节点详情
func (receiver node) Info(nodeName string) DockerNodeVO {
	progress := make(chan string, 1000)
	// docker node inspect node_1 --pretty
	var exitCode = exec.RunShell(fmt.Sprintf("docker node inspect %s --pretty", nodeName), progress, nil, "", false)
	serviceList := collections.NewListFromChan(progress)
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
				vo.Label.Add(DockerLabelVO{
					Name:  strings.Split(content, "=")[0],
					Value: strings.Split(content, "=")[1],
				})
			}
		}
	})
	return vo
}
