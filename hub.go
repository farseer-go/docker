package docker

import (
	"strings"

	"github.com/farseer-go/utils/exec"
)

type hub struct {
	api *dockerAPI
}

// Login 登陆仓库(使用Docker CLI客户端)
func (receiver hub) Login(dockerHub string, loginName string, loginPwd string) exec.ShellWait {
	if loginName != "" && loginPwd != "" {
		args := []string{"login", "-u", loginName, "--password-stdin"}
		if loginHost := getLoginHost(dockerHub); loginHost != "" {
			args = append(args, loginHost)
		}
		return exec.RunShellInput("docker", args, map[string]string{"DOCKER_HOST": ""}, "", true, loginPwd)
	}

	return exec.NewExitShellWait(-1, "登陆名和密码不能为空")
}

func getLoginHost(dockerHub string) string {
	dockerHub = strings.TrimSpace(strings.TrimRight(dockerHub, "/"))
	if dockerHub == "" || !strings.ContainsAny(dockerHub, ".:") {
		return ""
	}
	if index := strings.Index(dockerHub, "/"); index > -1 {
		return dockerHub[:index]
	}
	return dockerHub
}
