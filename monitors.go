package adoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (client *DockerClient) StopMonitor(id int64) {
	client.monitorLock.Lock()
	defer client.monitorLock.Unlock()
	delete(client.monitors, id)
}

func (client *DockerClient) newMonitorItem() int64 {
	client.monitorLock.Lock()
	defer client.monitorLock.Unlock()

	var id int64
	for trial := 5; trial > 0; trial -= 1 {
		id = random.Int63()
		if _, ok := client.monitors[id]; !ok {
			client.monitors[id] = struct{}{}
			break
		}
	}
	// we have some change to conflict, but I think it's ok
	return id
}

type EventCallback func(event Event, err error)

func (client *DockerClient) MonitorEvents(filters string, callback EventCallback) int64 {
	v := url.Values{}
	if filters != "" {
		v.Set("filters", filters)
	}
	uri := "events"
	if len(v) > 0 {
		uri += "?" + v.Encode()
	}
	id := client.newMonitorItem()
	go client.monitorEvents(id, uri, callback)
	return id
}

// will be running inside a goroutine
func (client *DockerClient) monitorEvents(id int64, uri string, callback EventCallback) {
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		decoder := json.NewDecoder(resp.Body)
		client.monitorLock.RLock()
		_, toContinue := client.monitors[id]
		client.monitorLock.RUnlock()
		for toContinue {
			var event Event
			if err := decoder.Decode(&event); err != nil {
				return err
			}
			callback(event, nil)
			client.monitorLock.RLock()
			_, toContinue = client.monitors[id]
			client.monitorLock.RUnlock()
		}
		return nil
	})
	if err != nil && err != io.EOF {
		callback(Event{}, err)
	}
}

type StatsCallback func(stats Stats, err error)

func (client *DockerClient) MonitorStats(id string, callback StatsCallback) int64 {
	uri := fmt.Sprintf("containers/%s/stats", id)
	monitorId := client.newMonitorItem()
	go client.monitorStats(monitorId, uri, callback)
	return monitorId
}

func (client *DockerClient) monitorStats(id int64, uri string, callback StatsCallback) {
	err := client.sendRequestCallback("GET", uri, nil, nil, func(resp *http.Response) error {
		decoder := json.NewDecoder(resp.Body)
		client.monitorLock.RLock()
		_, toContinue := client.monitors[id]
		client.monitorLock.RUnlock()
		for toContinue {
			var stats Stats
			if err := decoder.Decode(&stats); err != nil {
				return err
			}
			callback(stats, nil)
			client.monitorLock.RLock()
			_, toContinue = client.monitors[id]
			client.monitorLock.RUnlock()
		}
		return nil
	})
	if err != nil && err != io.EOF {
		callback(Stats{}, err)
	}
}
