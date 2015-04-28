package adoc

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AuthConfig struct {
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

func (auth AuthConfig) Encode() string {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(auth)
	return base64.URLEncoding.EncodeToString(buffer.Bytes())
}

type Error struct {
	StatusCode int
	Status     string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Status)
}

const (
	kDefaultApiVersion = "v1.17"
	kDefaultTimeout    = 30
)

type DockerClient struct {
	daemonUrl  *url.URL
	httpClient *http.Client
	tlsConfig  *tls.Config
	apiVersion string
}

func NewDockerClient(daemonUrl string, tlsConfig *tls.Config, apiVersion ...string) (*DockerClient, error) {
	return NewDockerClientTimeout(daemonUrl, tlsConfig, time.Duration(kDefaultTimeout*time.Second), apiVersion...)
}

func NewDockerClientTimeout(daemonUrl string, tlsConfig *tls.Config, timeout time.Duration, apiVersion ...string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Scheme == "tcp" {
		if tlsConfig == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	httpClient := newHttpClient(u, tlsConfig, timeout)
	clientApiVersion := kDefaultApiVersion
	if len(apiVersion) > 0 && apiVersion[0] != "" {
		clientApiVersion = apiVersion[0]
		if !strings.HasPrefix(clientApiVersion, "v") {
			clientApiVersion = "v" + clientApiVersion
		}
	}
	return &DockerClient{
		daemonUrl:  u,
		httpClient: httpClient,
		tlsConfig:  tlsConfig,
		apiVersion: clientApiVersion,
	}, nil
}

type responseCallback func(resp *http.Response) error

func (client *DockerClient) sendRequestCallback(method string, path string, body []byte, headers map[string]string, callback responseCallback) error {
	b := bytes.NewBuffer(body)
	urlPath := fmt.Sprintf("%s/%s/%s", client.daemonUrl.String(), client.apiVersion, path)
	fmt.Println(urlPath)
	req, err := http.NewRequest(method, urlPath, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") && client.tlsConfig == nil {
			return fmt.Errorf("%v. Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}
		return err
	}
	if resp.StatusCode >= 400 {
		return Error{resp.StatusCode, resp.Status}
	}

	defer resp.Body.Close()
	return callback(resp)
}

func (client *DockerClient) sendRequest(method string, path string, body []byte, headers map[string]string) ([]byte, error) {
	var data []byte
	err := client.sendRequestCallback(method, path, body, headers, func(resp *http.Response) error {
		var cbErr error
		if data, cbErr = ioutil.ReadAll(resp.Body); cbErr != nil {
			return cbErr
		}
		return nil
	})
	return data, err
}

func newHttpClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	switch u.Scheme {
	case "unix":
		socketPath := u.Path
		transport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	default:
		transport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	}
	return &http.Client{
		Transport: transport,
	}
}

// This part contains the misc apis listed in
// https://docs.docker.com/reference/api/docker_remote_api_v1.17/#23-misc

type Version struct {
	ApiVersion    string
	Arch          string
	GitCommit     string
	GoVersion     string
	KernelVersion string
	Os            string
	Version       string
}

type DockerInfo struct {
	Containers         int64
	Debug              int
	DockerRootDir      string
	Driver             string
	DriverStatus       [][]string
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
	SystemTime         time.Time
}

func (client *DockerClient) Version() (Version, error) {
	var ret Version
	if data, err := client.sendRequest("GET", "version", nil, nil); err != nil {
		return Version{}, err
	} else {
		if err := json.Unmarshal(data, &ret); err != nil {
			return ret, err
		}
		return ret, nil
	}
}

func (client *DockerClient) Info() (DockerInfo, error) {
	var ret DockerInfo
	if data, err := client.sendRequest("GET", "info", nil, nil); err != nil {
		return ret, err
	} else {
		if err := json.Unmarshal(data, &ret); err != nil {
			return ret, err
		}
		return ret, nil
	}
}

func (client *DockerClient) Ping() (bool, error) {
	if data, err := client.sendRequest("GET", "_ping", nil, nil); err != nil {
		return false, err
	} else {
		return string(data) == "OK", nil
	}
}

// Missing apis for
// auth
// commit: Create a new image from a container's changes
// events: Monitor Docker's events
// images/(name)/get: Get a tarball containing all images in a repository
// images/get: Get a tarball containing all images.
// images/load: Load a tarball with a set of images and tags into docker
// containers/(id)/exec: Exec Create
// exec/(id)/start: Exec Start
// exec/(id)/resize
// exec/(id)/json
