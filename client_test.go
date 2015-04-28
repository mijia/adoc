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
	fmt.Println("\n\n\n")
	containerConf := ContainerConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"date"},
		Image:        "busybox",
	}
	id, err := docker.CreateContainer(containerConf, HostConfig{})
	if err != nil {
		t.Fatalf("Cannot create the container, %s", err)
	}
	if err := docker.StartContainer(id); err != nil {
		t.Fatalf("Cannot start the container, %s", err)
	}
	containers, err := docker.ListContainers(true, true, "")
	if err != nil {
		t.Fatalf("Cannot list containers, %s", err)
	}
	d("Containers", containers)

	if err := docker.RenameContainer(id, "hello_renamed"); err != nil {
		t.Fatalf("Cannot rename the container, %s", err)
	}

	container, err := docker.InspectContainer(id)
	if err != nil {
		t.Fatalf("Cannot inpsect container, id=%s, %s", id, err)
	}
	d("Container", container)
	if container.Name != "/hello_renamed" {
		t.Fatalf("Rename failed, need=%s, but got=%s", "hello_renamed", container.Name)
	}

	if changes, err := docker.ContainerChanges(id); err != nil {
		t.Fatalf("Cannot get container changes, %s", err)
	} else {
		d("Container Changes", changes)
	}

	if logs, err := docker.ContainerLogs(id, true, true, false); err != nil {
		t.Fatalf("Cannot get logs from the container, %s", err)
	} else {
		d("Container Logs", logs)
	}

	docker.RemoveContainer(id, true, true)
}

func TestContainerCtls(t *testing.T) {
	fmt.Println("\n\n\n")
	containerConf := ContainerConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"python", "app.py"},
		Image:        "training/webapp",
	}
	hostConf := HostConfig{
		PortBindings: map[string][]PortBinding{
			"5000/tcp": []PortBinding{
				PortBinding{},
			},
		},
	}
	id, err := docker.CreateContainer(containerConf, hostConf, "test_container")
	if err != nil {
		t.Fatalf("Cannot create the container, %s", err)
	}
	d("Created Container ID", id)

	if err := docker.StartContainer(id); err != nil {
		t.Fatalf("Cannot start the container, %s", err)
	}

	if procs, err := docker.ContainerProcesses(id); err != nil {
		t.Fatalf("Cannot get top processes from the container, %s", err)
	} else {
		d("Processes", procs)
	}
	if stats, err := docker.ContainerStats(id); err != nil {
		t.Fatalf("Cannot get stats from the container, %s", err)
	} else {
		d("Container Stats", stats)
	}
	if err := docker.PauseContainer(id); err != nil {
		t.Fatalf("Cannot pause the container, %s", err)
	}
	if err := docker.UnpauseContainer(id); err != nil {
		t.Fatalf("Cannot unpause the container, %s", err)
	}

	//if code, err := docker.WaitContainer(id); err != nil {
	//t.Fatalf("Cannot wait on the container, %s", err)
	//} else {
	//d("Container Return", code)
	//}

	if err := docker.StopContainer(id); err != nil {
		t.Fatalf("Cannot stop the container, %s", err)
	}
	if err := docker.RestartContainer(id, 5); err != nil {
		t.Fatalf("Cannot restart the container, %s", err)
	}
	if err := docker.KillContainer(id); err != nil {
		t.Fatalf("Cannot kill the container, %s", err)
	}
	if err := docker.RemoveContainer(id, true, false); err != nil {
		t.Fatalf("Cannot remove the container, %s", err)
	}
}

func TestImages(t *testing.T) {
	fmt.Println("\n\n\n")
	if images, err := docker.ListImages(false); err != nil {
		t.Fatalf("Cannot list all the images, %s", err)
	} else {
		d("Images", images)
	}
	if image, err := docker.InspectImage("busybox"); err != nil {
		t.Fatalf("Cannot inspect image busybox, %s", err)
	} else {
		d("Image Detail", image)
	}

	if err := docker.RemoveImage("busybox", false, false); err != nil {
		t.Fatalf("Cannot remove the image, %s", err)
	}
	if err := docker.PullImage("busybox", "latest"); err != nil {
		t.Fatalf("Cannot pull the image, %s", err)
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
