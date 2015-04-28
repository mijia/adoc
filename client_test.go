package adoc

import (
	"fmt"
	"testing"
)

func TestVersionAndInfo(t *testing.T) {
	version, err := docker.Version()
	if err != nil {
		t.Fatalf("Cannot get docker version, %s", err)
	}
	d("ApiVersion", version)

	info, err := docker.Info()
	if err != nil {
		t.Fatalf("Cannot get docker info, %s", err)
	}
	d("DockerInfo", info)

	pong, err := docker.Ping()
	if err != nil || !pong {
		t.Fatalf("Cannot ping the docker, %s", err)
	}
}

func TestContainers(t *testing.T) {
	containers, err := docker.ListContainers(true, true, "")
	if err != nil {
		t.Fatalf("Cannot list containers, %s", err)
	}
	d("Containers", containers)

	if len(containers) > 0 {
		id := containers[0].Id
		container, err := docker.InspectContainer(id)
		if err != nil {
			t.Fatalf("Cannot inpsect container, id=%s, %s", id, err)
		}
		d("Container", container)
	}
}

func d(msg string, o interface{}) {
	fmt.Println(msg)
	fmt.Printf("%+v\n", o)
}

var docker *DockerClient

func init() {
	docker, _ = NewDockerClient("tcp://192.168.51.2:2375", nil)
}
