package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/utils/exec"
	"strings"
)

// Client docker client
type Client struct {
	dockerClient *client.Client
	Container    container
	Service      service
}

// NewClient 实例化一个Client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{
		dockerClient: cli,
		Container:    container{dockerClient: cli},
		Service:      service{dockerClient: cli},
	}, nil
}

func (receiver Client) GetVersion() string {
	version, err := receiver.dockerClient.ServerVersion(context.Background())
	if err != nil {
		flog.Warning(err.Error())
		return ""
	}
	return version.Version
}

func (receiver Client) Login(dockerHub string, loginName string, loginPwd string) error {
	if loginName != "" && loginPwd != "" {
		// 不包含域名的，意味着是登陆docker官网，不需要额外设置登陆的URL
		if !strings.Contains(dockerHub, ".") {
			dockerHub = ""
		}

		c := make(chan string, 100)
		var result = exec.RunShell("docker login "+dockerHub+" -u "+loginName+" -p "+loginPwd, c, nil, "", true)
		if result != 0 {
			return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
		}
	}
	return nil
}

func (receiver Client) Pull(image string) error {
	c := make(chan string, 100)
	exitCode := exec.RunShell(fmt.Sprintf("docker pull %s", image), c, nil, "", true)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}

// ClearImages 清除镜像
func (receiver Client) ClearImages() error {
	c := make(chan string, 100)
	var exitCode = exec.RunShell(`docker rmi $(docker images -f "dangling=true" -q) && docker builder prune -f && docker system prune -f`, c, nil, "", false)
	if exitCode != 0 {
		return fmt.Errorf(collections.NewListFromChan(c).ToString("\n"))
	}
	return nil
}
