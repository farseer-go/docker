package docker

import (
	"fmt"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/utils/exec"
)

type hub struct {
	progress chan string
}

// Login 登陆仓库
func (receiver hub) Login(dockerHub string, loginName string, loginPwd string) error {
	if loginName != "" && loginPwd != "" {
		// 不包含域名的，意味着是登陆docker官网，不需要额外设置登陆的URL
		if !strings.Contains(dockerHub, ".") {
			dockerHub = ""
		}

		var result = exec.RunShell("docker login "+dockerHub+" -u "+loginName+" -p "+loginPwd, receiver.progress, nil, "", true)
		if result != 0 {
			return fmt.Errorf(collections.NewListFromChan(receiver.progress).ToString("\n"))
		}
	}
	return nil
}
