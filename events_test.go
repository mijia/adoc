package adoc

import (
	"fmt"
	"testing"
	"time"
)

func TestEventsPolling(t *testing.T) {
	events, err := docker.EventsSince("", time.Hour, 5*time.Minute)
	if err != nil {
		t.Fatalf("Cannot poll the events, %s", err)
	}
	d("Events", events)
}

func TestEventsMonitor(t *testing.T) {
	monitorId := docker.MonitorEvents("", func(event Event, err error) {
		if err != nil {
			t.Errorf("Error when calling monitor, %s", err)
		} else {
			d("Event", event)
		}
	})
	fmt.Println("Please do some docker operations")
	time.Sleep(2 * time.Minute)
	docker.StopMonitor(monitorId)
}

func TestStatsMonitor(t *testing.T) {
	containerConf := ContainerConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"python", "app.py"},
		Image:        "training/webapp",
	}
	id, err := docker.CreateContainer(containerConf, HostConfig{})
	if err != nil {
		t.Fatalf("Cannot create the container, %s", err)
	}
	if err := docker.StartContainer(id); err != nil {
		t.Fatalf("Cannot start the container, %s", err)
	}

	monitorId := docker.MonitorStats(id, func(stats Stats, err error) {
		if err != nil {
			t.Errorf("Error when calling monitor, %s", err)
		} else {
			d("Stats CpuUsage.PercpuUsage", stats.CpuStats.CpuUsage.PercpuUsage)
		}
	})
	time.Sleep(30 * time.Second)
	docker.StopMonitor(monitorId)
	docker.RemoveContainer(id, true, true)
}
