package docker

import (
	"strings"

	"github.com/farseer-go/utils/exec"
)

type hub struct {
	//progress chan string
}

// Login 登陆仓库
func (receiver hub) Login(dockerHub string, loginName string, loginPwd string) (chan string, func() int) {
	if loginName != "" && loginPwd != "" {
		// 不包含域名的，意味着是登陆docker官网，不需要额外设置登陆的URL
		if !strings.Contains(dockerHub, ".") {
			dockerHub = ""
		}

		return exec.RunShell("docker login "+dockerHub+" -u "+loginName+" -p "+loginPwd, nil, "", true)
	}

	result := make(chan string, 1)
	result <- "登陆名和密码不能为空"
	close(result)
	return result, func() int { return -1 }
}
