package adoc

import (
	"encoding/json"
	"fmt"
	"time"
)

// This part contains the misc apis listed in
// https://docs.docker.com/reference/api/docker_remote_api_v1.17/#23-misc

type Version struct {
	ApiVersion    string
	GitCommit     string
	GoVersion     string
	Version       string
	Os            string // v1.18
	Arch          string // v1.18
	KernelVersion string // v1.18
}

type DockerInfo struct {
	Containers         int64
	DockerRootDir      string
	Driver             string
	DriverStatus       [][2]string
	ExecutionDriver    string
	ID                 string
	IPv4Forwarding     int
	Images             int64
	IndexServerAddress string
	InitPath           string
	InitSha1           string
	KernelVersion      string
	Labels             []string
	MemTotal           int64
	MemoryLimit        int
	NCPU               int64
	NEventsListener    int64
	NFd                int64
	NGoroutines        int64
	Name               string
	OperatingSystem    string
	SwapLimit          int
	HttpProxy          string    // v1.18
	HttpsProxy         string    // v1.18
	NoProxy            string    // v1.18
	SystemTime         time.Time // v1.18
	//Debug              bool // this will conflict with docker api and swarm api, fuck
}

type ExecConfig struct {
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
	Cmd          []string
}

func (client *DockerClient) Version() (Version, error) {
	var ret Version
	if data, err := client.sendRequest("GET", "version", nil, nil); err != nil {
		return Version{}, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) Info() (DockerInfo, error) {
	var ret DockerInfo
	if data, err := client.sendRequest("GET", "info", nil, nil); err != nil {
		return ret, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) Ping() (bool, error) {
	if data, err := client.sendRequest("GET", "_ping", nil, nil); err != nil {
		return false, err
	} else {
		return string(data) == "OK", nil
	}
}

func (client *DockerClient) CreateExec(id string, execConfig ExecConfig) (string, error) {
	if body, err := json.Marshal(execConfig); err != nil {
		return "", err
	} else {
		uri := fmt.Sprintf("containers/%s/exec", id)
		if data, err := client.sendRequest("POST", uri, body, nil); err != nil {
			return "", err
		} else {
			var ret map[string]interface{}
			if err := json.Unmarshal(data, &ret); err != nil {
				return "", err
			}
			if execId, ok := ret["Id"]; ok {
				return execId.(string), nil
			}
			return "", fmt.Errorf("Cannot find Id field inside result object, %+v", ret)
		}
	}
}

func (client *DockerClient) StartExec(execId string, detach, tty bool) ([]byte, error) {
	params := map[string]bool{
		"Detach": detach,
		"Tty":    tty,
	}
	if body, err := json.Marshal(params); err != nil {
		return nil, err
	} else {
		uri := fmt.Sprintf("exec/%s/start", execId)
		return client.sendRequest("POST", uri, body, nil)
	}
}

// Missing apis for
// auth
// commit: Create a new image from a container's changes
// events: Monitor Docker's events
// images/(name)/get: Get a tarball containing all images in a repository
// images/get: Get a tarball containing all images.
// images/load: Load a tarball with a set of images and tags into docker
// exec/(id)/resize
// exec/(id)/json
