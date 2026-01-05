package client

import (
	"encoding/json"
)

// Monitor represents a Ceph monitor
type Monitor struct {
	Name       string `json:"name"`
	Rank       int    `json:"rank"`
	Addr       string `json:"addr"`
	PublicAddr string `json:"public_addr"`
}

// MonMap represents the monitor map
type MonMap struct {
	Fsid string    `json:"fsid"`
	Mons []Monitor `json:"mons"`
}

// MonStatus represents the monitor status
type MonStatus struct {
	MonMap MonMap `json:"monmap"`
}

// MonitorResponse represents the response from /api/monitor
type MonitorResponse struct {
	MonStatus MonStatus `json:"mon_status"`
}

// GetClusterFSID retrieves the cluster FSID
func (c *Client) GetClusterFSID() (string, error) {
	resp, err := c.DoRequest("GET", "/api/monitor", nil)
	if err != nil {
		return "", err
	}

	var response MonitorResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return "", err
	}

	return response.MonStatus.MonMap.Fsid, nil
}

// GetMonitors retrieves the list of monitors
func (c *Client) GetMonitors() ([]Monitor, error) {
	resp, err := c.DoRequest("GET", "/api/monitor", nil)
	if err != nil {
		return nil, err
	}

	var response MonitorResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}

	return response.MonStatus.MonMap.Mons, nil
}
