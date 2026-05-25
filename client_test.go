package docker

import "testing"

func TestParseDockerHost(t *testing.T) {
	t.Setenv("DOCKER_TLS_VERIFY", "")

	tests := []struct {
		name    string
		rawHost string
		want    dockerEndpoint
	}{
		{
			name:    "default unix socket",
			rawHost: "",
			want: dockerEndpoint{
				rawHost: defaultDockerHost,
				scheme:  "unix",
				address: "/var/run/docker.sock",
				baseURL: "http://docker",
			},
		},
		{
			name:    "explicit unix socket",
			rawHost: "unix:///var/run/docker.sock",
			want: dockerEndpoint{
				rawHost: "unix:///var/run/docker.sock",
				scheme:  "unix",
				address: "/var/run/docker.sock",
				baseURL: "http://docker",
			},
		},
		{
			name:    "plain tcp docker host",
			rawHost: "tcp://172.19.235.16:2375",
			want: dockerEndpoint{
				rawHost: "tcp://172.19.235.16:2375",
				scheme:  "tcp",
				address: "172.19.235.16:2375",
				baseURL: "http://172.19.235.16:2375",
			},
		},
		{
			name:    "http docker host",
			rawHost: "http://172.19.235.16:2375",
			want: dockerEndpoint{
				rawHost: "http://172.19.235.16:2375",
				scheme:  "tcp",
				address: "172.19.235.16:2375",
				baseURL: "http://172.19.235.16:2375",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDockerHost(tt.rawHost)
			if err != nil {
				t.Fatalf("parseDockerHost() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseDockerHost() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseDockerHostUnsupportedScheme(t *testing.T) {
	t.Setenv("DOCKER_TLS_VERIFY", "")

	_, err := parseDockerHost("ssh://docker-host")
	if err == nil {
		t.Fatal("parseDockerHost() expected unsupported scheme error")
	}
}

func TestParseDockerHostUnsupportedTLS(t *testing.T) {
	t.Setenv("DOCKER_TLS_VERIFY", "1")

	_, err := parseDockerHost("tcp://172.19.235.16:2376")
	if err == nil {
		t.Fatal("parseDockerHost() expected unsupported TLS error")
	}
}

func TestDockerAPIURL(t *testing.T) {
	api := dockerAPI{endpoint: dockerEndpoint{baseURL: "http://172.19.235.16:2375"}}
	if got := api.URL("/version"); got != "http://172.19.235.16:2375/version" {
		t.Fatalf("URL() = %q", got)
	}

	api = dockerAPI{endpoint: dockerEndpoint{baseURL: "http://docker/"}}
	if got := api.URL("version"); got != "http://docker/version" {
		t.Fatalf("URL() = %q", got)
	}
}
