package docker

import (
	"encoding/json"
	"net/http"
)

// UnixGet 通过Unix Socket发送HTTP GET请求，并将响应解析为指定类型
func UnixGet[T any](unixClient *http.Client, url string) (T, error) {
	var t T
	resp, err := unixClient.Get(url)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}
