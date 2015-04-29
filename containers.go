package adoc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// This part contains apis for the containers listed in
// https://docs.docker.com/reference/api/docker_remote_api_v1.17/#21-containers

type Port struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

type Container struct {
	Command    string
	Created    int64
	Id         string
	Image      string
	Labels     map[string]string
	Names      []string
	Ports      []Port
	SizeRootFs int64
	SizeRw     int64
	Status     string
}

type ContainerConfig struct {
	AttachStderr    bool
	AttachStdin     bool
	AttachStdout    bool
	Cmd             []string
	CpuShares       int
	Cpuset          string
	Domainname      string
	Entrypoint      []string
	Env             []string
	ExposedPorts    map[string]struct{}
	Hostname        string
	Image           string
	Labels          map[string]string
	MacAddress      string
	Memory          int64
	MemorySwap      int64
	NetworkDisabled bool
	OnBuild         []string
	OpenStdin       bool
	PortSpecs       []string
	StdinOnce       bool
	Tty             bool
	User            string
	Volumes         map[string]struct{}
	WorkingDir      string
}

type Device struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}

type RestartPolicy struct {
	MaximumRetryCount int
	Name              string
}

type Ulimit struct {
	Name string
	Soft int64
	Hard int64
}

type LogConfig struct {
	Type   string
	Config map[string]string
}

type HostConfig struct {
	Binds           []string
	CapAdd          []string
	CapDrop         []string
	CgroupParent    string
	ContainerIDFile string
	CpuShares       int
	CpusetCpus      string
	Devices         []Device
	Dns             []string
	DnsSearch       []string
	ExtraHosts      []string
	IpcMode         string
	Links           []string
	LxcConf         []map[string]string
	Memory          int64
	MemorySwap      int64
	NetworkMode     string
	PidMode         string
	PortBindings    map[string][]PortBinding
	Privileged      bool
	PublishAllPorts bool
	ReadonlyRootfs  bool
	RestartPolicy   RestartPolicy
	SecurityOpt     []string
	VolumesFrom     []string
	Ulimits         []Ulimit  // 1.18
	LogConfig       LogConfig // 1.18
}

type PortBinding struct {
	HostIp   string
	HostPort string
}

type NetworkSettings struct {
	Bridge                 string
	Gateway                string
	GlobalIPv6Address      string
	GlobalIPv6PrefixLen    int
	IPAddress              string
	IPPrefixLen            int
	IPv6Gateway            string
	LinkLocalIPv6Address   string
	LinkLocalIPv6PrefixLen int
	MacAddress             string
	Ports                  map[string][]PortBinding
	// PortMapping: null,
}

type ContainerState struct {
	Dead       bool
	Error      string
	ExitCode   int
	FinishedAt time.Time
	OOMKilled  bool
	Paused     bool
	Pid        int64
	Restarting bool
	Running    bool
	StartedAt  time.Time
}

type ContainerDetail struct {
	AppArmorProfile string
	Args            []string
	Config          ContainerConfig
	Created         time.Time
	Driver          string
	ExecDriver      string
	ExecIDs         []string
	HostConfig      HostConfig
	HostnamePath    string
	HostsPath       string
	Id              string
	Image           string
	LogPath         string
	MountLabel      string
	Name            string
	NetworkSettings NetworkSettings
	Path            string
	ProcessLabel    string
	ResolvConfPath  string
	RestartCount    int
	State           ContainerState
	Volumes         map[string]string
	VolumesRW       map[string]string
}

func (client *DockerClient) ListContainers(showAll, showSize bool, filters ...string) ([]Container, error) {
	v := url.Values{}
	v.Set("all", strconv.FormatBool(showAll))
	v.Set("size", strconv.FormatBool(showSize))
	if len(filters) > 0 && filters[0] != "" {
		v.Set("filters", filters[0])
	}
	uri := fmt.Sprintf("containers/json?%s", v.Encode())
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return nil, err
	} else {
		var ret []Container
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) InspectContainer(id string) (ContainerDetail, error) {
	uri := fmt.Sprintf("containers/%s/json", id)
	var ret ContainerDetail
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return ret, err
	} else {
		err := json.Unmarshal(data, &ret)
		return ret, err
	}
}

func (client *DockerClient) CreateContainer(containerConf ContainerConfig, hostConf HostConfig, name ...string) (string, error) {
	var config struct {
		ContainerConfig
		HostConfig HostConfig
	}
	config.ContainerConfig = containerConf
	config.HostConfig = hostConf

	if body, err := json.Marshal(config); err != nil {
		return "", err
	} else {
		uri := "containers/create"
		if len(name) > 0 && name[0] != "" {
			v := url.Values{}
			v.Set("name", name[0])
			uri += "?" + v.Encode()
		}
		if data, err := client.sendRequest("POST", uri, body, nil); err != nil {
			return "", err
		} else {
			var resp struct {
				Id       string
				Warnings []string
			}
			err := json.Unmarshal(data, &resp)
			if len(resp.Warnings) > 0 {
				logger.Warnf("Create container returns warning from docker daemon: %+v", resp.Warnings)
			}
			return resp.Id, err
		}
	}
}

func (client *DockerClient) StartContainer(id string) error {
	uri := fmt.Sprintf("containers/%s/start", id)
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) StopContainer(id string, timeout ...int) error {
	uri := fmt.Sprintf("containers/%s/stop", id)
	if len(timeout) > 0 && timeout[0] >= 0 {
		v := url.Values{}
		v.Set("t", fmt.Sprintf("%d", timeout[0]))
		uri += "?" + v.Encode()
	}
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) RestartContainer(id string, timeout ...int) error {
	uri := fmt.Sprintf("containers/%s/restart", id)
	if len(timeout) > 0 && timeout[0] >= 0 {
		v := url.Values{}
		v.Set("t", fmt.Sprintf("%d", timeout[0]))
		uri += "?" + v.Encode()
	}
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) KillContainer(id string, signal ...string) error {
	uri := fmt.Sprintf("containers/%s/kill", id)
	if len(signal) > 0 && signal[0] != "" {
		v := url.Values{}
		v.Set("signal", signal[0])
		uri += "?" + v.Encode()
	}
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) PauseContainer(id string) error {
	uri := fmt.Sprintf("containers/%s/pause", id)
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) UnpauseContainer(id string) error {
	uri := fmt.Sprintf("containers/%s/unpause", id)
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

func (client *DockerClient) RemoveContainer(id string, force, volumes bool) error {
	v := url.Values{}
	v.Set("force", strconv.FormatBool(force))
	v.Set("v", strconv.FormatBool(volumes))
	uri := fmt.Sprintf("containers/%s?%s", id, v.Encode())
	_, err := client.sendRequest("DELETE", uri, nil, nil)
	return err
}

func (client *DockerClient) RenameContainer(id string, name string) error {
	v := url.Values{}
	v.Set("name", name)
	uri := fmt.Sprintf("containers/%s/rename?%s", id, v.Encode())
	_, err := client.sendRequest("POST", uri, nil, nil)
	return err
}

// This will block the call routine until the container is stopped
func (client *DockerClient) WaitContainer(id string) (int, error) {
	uri := fmt.Sprintf("containers/%s/wait", id)
	if data, err := client.sendRequest("POST", uri, nil, nil); err != nil {
		return 0, err
	} else {
		var ret map[string]int
		if err := json.Unmarshal(data, &ret); err != nil {
			return 0, err
		}
		if code, ok := ret["StatusCode"]; ok {
			return code, nil
		} else {
			logger.Warnf("There is no StatusCode key inside results map, the API maybe changed, ret=%+v", ret)
			return 0, fmt.Errorf("Cannot get StatusCode from return data, %+v", ret)
		}
	}
}

func (client *DockerClient) ContainerLogs(id string, stdout, stderr, timestamps bool, tail ...int) ([]LogEntry, error) {
	// no following mode
	v := url.Values{}
	v.Set("stdout", strconv.FormatBool(stdout))
	v.Set("stderr", strconv.FormatBool(stderr))
	v.Set("timestamps", strconv.FormatBool(timestamps))
	if len(tail) > 0 && tail[0] >= 0 {
		v.Set("tail", fmt.Sprintf("%d", tail[0]))
	}
	uri := fmt.Sprintf("containers/%s/logs?%s", id, v.Encode())

	var entries []LogEntry
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		var cbErr error
		entries, cbErr = ReadAllDockerLogs(resp.Body)
		return cbErr
	})
	return entries, err
}

type Processes struct {
	Titles    []string
	Processes [][]string
}

func (client *DockerClient) ContainerProcesses(id string, psArgs ...string) (Processes, error) {
	var procs Processes
	v := url.Values{}
	if len(psArgs) > 0 && psArgs[0] != "" {
		v.Set("ps_args", psArgs[0])
	}
	uri := fmt.Sprintf("containers/%s/top", id)
	if len(v) > 0 {
		uri += "?" + v.Encode()
	}
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return procs, err
	} else {
		err := json.Unmarshal(data, &procs)
		return procs, err
	}
}

type FsChange struct {
	Path string
	Kind int
}

func (client *DockerClient) ContainerChanges(id string) ([]FsChange, error) {
	uri := fmt.Sprintf("containers/%s/changes", id)
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return nil, err
	} else {
		var changes []FsChange
		err := json.Unmarshal(data, &changes)
		return changes, err
	}
}

// Missing apis for
// containers/(id)/copy
// containers/(id)/attach
// containers/(id)/export
// containers/(id)/resize?h=<height>&w=<width>
// containers/(id)/attach/ws
