package docker

import (
	"net/http"
	"net/url"

	"fmt"
)

type task struct {
	unixClient *http.Client
}

// Inspect 查看服务详情
// 注意：由于 Service 本身不包含 ContainerID，此方法实际上是通过 Service ID 查询其关联的 Task 列表
func (receiver task) Inspect(taskId string) (ServiceIdInspectJson, error) {
	// 1. 构造 URL
	// 使用 /tasks 接口，过滤出属于该 Service 的任务
	// 这将返回 Task 对象列表，其中包含 ContainerID
	filter := fmt.Sprintf(`{"service":{"%s":true}}`, taskId)
	url := "http://localhost/tasks?filters=" + url.QueryEscape(filter)

	// 2. 发送请求并解析
	// ServiceIdInspectJson 应该是 []Task 结构，因为返回的是数组
	results, err := UnixGetDecode[[]ServiceIdInspectJson](receiver.unixClient, url)
	if err != nil {
		return ServiceIdInspectJson{}, err
	}

	// 3. 如果没有找到任务（服务不存在或无副本）
	if len(results) == 0 {
		return ServiceIdInspectJson{}, nil
	}
	var result ServiceIdInspectJson
	// 4. 处理 ContainerID 截断 (保持原逻辑)
	// 注意：这里假设 ServiceIdInspectJson 是切片类型，且元素结构包含 Status.ContainerStatus
	// 由于 Go 泛型或类型断言无法直接访问字段，此处逻辑需根据您的具体结构体定义调整。
	// 假设 result 是切片：
	if len(results) > 0 {
		result = results[0] // 取第一个任务的详情，实际情况可能需要根据业务逻辑处理多个任务

		// 这里需要根据您的 ServiceIdInspectJson 具体定义来写
		// 假设它定义类似: type ServiceIdInspectJson []TaskStruct
		if len(result.Status.ContainerStatus.ContainerID) >= 12 {
			result.Status.ContainerStatus.ContainerID = result.Status.ContainerStatus.ContainerID[:12]
		}
	}

	return result, nil
}
