package test

import (
	"github.com/farseer-go/docker"
	"testing"
)

func TestStats(t *testing.T) {
	docker.NewClient().Stats()
}
