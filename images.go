package docker

import (
	"fmt"
	"net/http"

	"github.com/farseer-go/utils/exec"
)

type images struct {
	unixClient *http.Client
}

// 4. 解析 JSON
type PruneResult struct {
	ImagesDeleted []struct {
		Untagged string `json:"Untagged"`
		Deleted  string `json:"Deleted"`
	} `json:"ImagesDeleted"`
	SpaceReclaimed int64 `json:"SpaceReclaimed"`
}

// Pull 拉取镜像(使用Docker CLI客户端)
func (receiver images) Pull(image string) (chan string, func() int) {
	return exec.RunShell(fmt.Sprintf("docker pull %s", image), nil, "", true)
}

// ClearImages 清除镜像
func (receiver images) ClearImages() ([]string, error) {
	url := "http://localhost/system/prune?all=true&force=true"

	// 一行代码搞定请求和解析
	result, err := UnixPostDecode[PruneResult](receiver.unixClient, url)
	if err != nil {
		return nil, err
	}

	// 组装日志列表
	var logs []string
	for _, img := range result.ImagesDeleted {
		if img.Untagged != "" {
			logs = append(logs, fmt.Sprintf("Untagged: %s", img.Untagged))
		}
		if img.Deleted != "" {
			logs = append(logs, fmt.Sprintf("Deleted: %s", img.Deleted))
		}
	}

	mb := float64(result.SpaceReclaimed) / 1024 / 1024
	logs = append(logs, fmt.Sprintf("Total reclaimed space: %.2f MB", mb))

	return logs, nil
}
