package adoc

import (
	"encoding/json"
	"fmt"
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

func (client *DockerClient) ListContainers(showAll, showSize bool, filters string) ([]Container, error) {
	all, size := 0, 0
	if showAll {
		all = 1
	}
	if showSize {
		size = 1
	}
	uri := fmt.Sprintf("containers/json?all=%d&size=%d", all, size)
	if filters != "" {
		uri += "&filters=" + filters
	}

	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return nil, err
	} else {
		var ret []Container
		if err := json.Unmarshal(data, &ret); err != nil {
			return nil, err
		}
		return ret, nil
	}
}

// CreateContainer

func (client *DockerClient) InspectContainer(id string) (ContainerDetail, error) {
	uri := fmt.Sprintf("containers/%s/json", id)

	var ret ContainerDetail
	if data, err := client.sendRequest("GET", uri, nil, nil); err != nil {
		return ret, err
	} else {
		if err := json.Unmarshal(data, &ret); err != nil {
			return ret, err
		}
		return ret, nil
	}
}
