package adoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DockerEventCreate      = "create"
	DockerEventDestroy     = "destroy"
	DockerEventDie         = "die"
	DockerEventExecCreate  = "exec_create"
	DockerEventExecStart   = "exec_start"
	DockerEventExport      = "export"
	DockerEventKill        = "kill"
	DockerEventOOM         = "oom"
	DockerEventPause       = "pause"
	DockerEventRestart     = "restart"
	DockerEventStart       = "start"
	DockerEventStop        = "stop"
	DockerEventUnpause     = "unpause"
	DockerEventRename      = "rename"
	DockerEventImageUntag  = "untag"
	DockerEventImageDelete = "delete"

	// ContainerEventType is the event type that containers generate
	ContainerEventType = "container"
	// DaemonEventType is the event type that daemon generate
	DaemonEventType = "daemon"
	// ImageEventType is the event type that images generate
	ImageEventType = "image"
	// NetworkEventType is the event type that networks generate
	NetworkEventType = "network"
	// PluginEventType is the event type that plugins generate
	PluginEventType = "plugin"
	// VolumeEventType is the event type that volumes generate
	VolumeEventType = "volume"
)

// Actor describes something that generates events,
// like a container, or a network, or a volume.
// It has a defined name and a set or attributes.
// The container attributes are its labels, other actors
// can generate these attributes from other properties.
type Actor struct {
	ID         string
	Attributes map[string]string
}

// Message represents the information an event contains
type Event struct {
	// Deprecated information from JSONMessage.
	// With data only in container events.
	Status string `json:"status,omitempty"`
	ID     string `json:"id,omitempty"`
	From   string `json:"from,omitempty"`

	Type   string
	Action string
	Actor  Actor

	Time     int64 `json:"time,omitempty"`
	TimeNano int64 `json:"timeNano,omitempty"`

	Node SwarmNode `json:"node,omitempty"`
}

// type Event struct {
// 	Id     string    `json:"id"`
// 	Status string    `json:"status"`
// 	From   string    `json:"from"`
// 	Time   int64     `json:"time"`
// 	Node   SwarmNode `json:"node,omitempty"`
// }

func (client *DockerClient) EventsSince(filters string, since time.Duration, until ...time.Duration) ([]Event, error) {
	if client.isSwarm {
		return nil, fmt.Errorf("Swarm doesn't support the events polling mode.")
	}

	v := url.Values{}
	if filters != "" {
		v.Set("filters", filters)
	}
	now := time.Now()
	v.Set("since", fmt.Sprintf("%d", tsFromNow(now, since)))
	if len(until) > 0 {
		v.Set("until", fmt.Sprintf("%d", tsFromNow(now, until[0])))
	}
	uri := fmt.Sprintf("events?%s", v.Encode())

	events := make([]Event, 0)
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		var event Event
		var cbErr error
		decoder := json.NewDecoder(resp.Body)
		for ; cbErr == nil; cbErr = decoder.Decode(&event) {
			// Not sure about why there will be an empty event first
			if cbErr == nil && event.Status != "" {
				events = append(events, event)
			}
		}
		if cbErr != io.EOF {
			return cbErr
		}
		return nil
	}, nil)
	return events, err
}

func tsFromNow(now time.Time, duration time.Duration) int64 {
	return now.Add(-1 * duration).Unix()
}
