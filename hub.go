package docker

import (
	"net/http"
	"strings"

	"github.com/farseer-go/utils/exec"
)

type hub struct {
	unixClient *http.Client
}

// Login 登陆仓库(使用Docker CLI客户端)
func (receiver hub) Login(dockerHub string, loginName string, loginPwd string) exec.ShellWait {
	if loginName != "" && loginPwd != "" {
		// 不包含域名的，意味着是登陆docker官网，不需要额外设置登陆的URL
		if !strings.Contains(dockerHub, ".") {
			dockerHub = ""
		}

		//return exec.RunShell("docker login "+dockerHub+" -u "+loginName+" -p "+loginPwd, nil, "", true)
		return exec.RunShell("docker", []string{"login", dockerHub, "-u", loginName, "-p", loginPwd}, nil, "", true)
	}

	return exec.NewExitShellWait(-1, "登陆名和密码不能为空")
}
