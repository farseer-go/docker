package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/farseer-go/collections"
)

type config struct {
	unixClient *http.Client
}
type ConfigCreateRequest struct {
	Name   string            `json:"Name"`
	Labels map[string]string `json:"Labels"`
	Data   string            `json:"Data"` // 注意：发送给 API 时必须是 Base64 编码的字符串
}

type ConfigInfo struct {
	ID   string              `json:"ID"`
	Spec ConfigCreateRequest `json:"Spec"`
}

// CreateConfig 创建一个新的 Docker Config
func (receiver config) Create(name string, content []byte, labels map[string]string) (string, error) {
	url := "http://localhost/configs/create"

	data := ConfigCreateRequest{
		Name:   name,
		Labels: labels,
		Data:   base64.StdEncoding.EncodeToString(content), // 必须 Base64
	}

	body, _ := json.Marshal(data)
	resp, err := receiver.unixClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create config failed: %d", resp.StatusCode)
	}

	var result struct{ ID string }
	json.NewDecoder(resp.Body).Decode(&result)
	return result.ID, nil
}

// Inspect 查看单个配置详情 (通过 ID 或 Name)
func (receiver config) Inspect(configIdOrName string) (ConfigInfo, error) {
	// 1. 构造 URL: /configs/{id_or_name}
	url := fmt.Sprintf("http://localhost/configs/%s", configIdOrName)

	// 2. 调用工具函数解析
	result, err := UnixGetDecode[ConfigInfo](receiver.unixClient, url)
	if err != nil {
		return result, err
	}

	// 3. 校验是否存在
	if result.ID == "" {
		return result, errors.New("no such config")
	}
	// 3. 将 Base64 的 Data 解码为明文 string
	decodedByte, err := base64.StdEncoding.DecodeString(result.Spec.Data)
	if err != nil {
		return result, fmt.Errorf("decode base64 failed: %v", err)
	}

	result.Spec.Data = string(decodedByte)
	return result, nil
}

// InspectByService 根据 Label 查找关联的所有配置
func (receiver config) InspectByService(serviceName string) (ConfigInfo, error) {
	// curl --unix-socket /var/run/docker.sock http://localhost/configs

	// 使用 filter 过滤 Label
	filter := fmt.Sprintf(`{"label": ["owner_service=%s"]}`, serviceName)
	url := fmt.Sprintf("http://localhost/configs?filters=%s", url.QueryEscape(filter))

	configs, err := UnixGetDecode[collections.List[ConfigInfo]](receiver.unixClient, url)
	result := configs.First()
	if err != nil {
		return result, err
	}

	// 3. 校验是否存在
	if result.ID == "" {
		return result, errors.New("no such config")
	}
	// 3. 将 Base64 的 Data 解码为明文 string
	decodedByte, err := base64.StdEncoding.DecodeString(result.Spec.Data)
	if err != nil {
		return result, fmt.Errorf("decode base64 failed: %v", err)
	}

	result.Spec.Data = string(decodedByte)
	return result, nil
}
