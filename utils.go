package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UnixGetDecode 通过Unix Socket发送HTTP请求，并将响应解析为指定类型
func UnixGetDecode[T any](unixClient *http.Client, url string) (T, error) {
	var t T
	resp, err := unixClient.Get(url)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

// UnixPostDecode 发送POST请求，并将响应解析为指定类型 (适用于 Prune 等返回数据的接口)
func UnixPostDecode[T any](unixClient *http.Client, url string) (T, error) {
	var t T
	resp, err := unixClient.Post(url, "application/json", nil)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	// Prune 接口成功返回 200 OK，而不是 204
	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&t)
		return t, nil
	}

	// 处理错误
	body, _ := io.ReadAll(resp.Body)
	return t, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
}

// UnixGet 通过UnixGet Socket发送HTTP GET请求，并将响应解析为指定类型
func UnixGet(unixClient *http.Client, url string) (*http.Response, error) {
	resp, err := unixClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}

// UnixPost 通过UnixGet Socket发送HTTP POST请求，并将响应解析为指定类型
func UnixPost(unixClient *http.Client, url string) (*http.Response, error) {
	// 发送 POST 请求 (Body 为 nil)
	resp, err := unixClient.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 204 No Content 表示成功
	if resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}

	// 解析错误信息
	body, _ := io.ReadAll(resp.Body)
	return resp, fmt.Errorf("kill failed (%d): %s", resp.StatusCode, string(body))
}

// UnixDelete 通过UnixGet Socket发送HTTP DELETE请求，并将响应解析为指定类型
func UnixDelete(unixClient *http.Client, url string) (*http.Response, error) {
	// 发送 DELETE 请求 (Body 为 nil)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := unixClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 204 No Content 表示成功
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return resp, nil
	}

	// 解析错误信息
	body, _ := io.ReadAll(resp.Body)
	return resp, fmt.Errorf("delete failed (%d): %s", resp.StatusCode, string(body))
}
