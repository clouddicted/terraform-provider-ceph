package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PoolConfiguration represents the configuration for a pool
type PoolConfiguration struct {
	RbdQosBpsLimit       int `json:"rbd_qos_bps_limit,omitempty"`
	RbdQosIopsLimit      int `json:"rbd_qos_iops_limit,omitempty"`
	RbdQosReadBpsLimit   int `json:"rbd_qos_read_bps_limit,omitempty"`
	RbdQosReadIopsLimit  int `json:"rbd_qos_read_iops_limit,omitempty"`
	RbdQosWriteBpsLimit  int `json:"rbd_qos_write_bps_limit,omitempty"`
	RbdQosWriteIopsLimit int `json:"rbd_qos_write_iops_limit,omitempty"`
	RbdQosBpsBurst       int `json:"rbd_qos_bps_burst,omitempty"`
	RbdQosIopsBurst      int `json:"rbd_qos_iops_burst,omitempty"`
	RbdQosReadBpsBurst   int `json:"rbd_qos_read_bps_burst,omitempty"`
	RbdQosReadIopsBurst  int `json:"rbd_qos_read_iops_burst,omitempty"`
	RbdQosWriteBpsBurst  int `json:"rbd_qos_write_bps_burst,omitempty"`
	RbdQosWriteIopsBurst int `json:"rbd_qos_write_iops_burst,omitempty"`
}

// Pool represents a Ceph pool
type Pool struct {
	PoolName            string            `json:"pool"` // Note: API expects "pool" for create, but might return "pool_name" in get? Let's check.
	Type                string            `json:"pool_type,omitempty"`
	PgAutoscaleMode     string            `json:"pg_autoscale_mode,omitempty"`
	PgNum               int               `json:"pg_num,omitempty"`
	Size                int               `json:"size,omitempty"`
	RuleName            string            `json:"rule_name,omitempty"`
	QuotaMaxBytes       int64             `json:"quota_max_bytes,omitempty"`
	ApplicationMetadata []string          `json:"application_metadata,omitempty"`
	RbdMirroring        bool              `json:"rbd_mirroring,omitempty"`
	Configuration       PoolConfiguration `json:"configuration,omitempty"`
}

// CreatePool creates a new pool
func (c *Client) CreatePool(pool Pool) error {
	rb, err := json.Marshal(pool)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("POST", "/api/pool", bytes.NewBuffer(rb))
	return err
}

// GetPool retrieves a pool by name
func (c *Client) GetPool(name string) (*Pool, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("/api/pool/%s", name), nil)
	if err != nil {
		return nil, err
	}

	// GET response has different field names than POST request
	type GetPoolResponse struct {
		PoolID              int      `json:"pool"`
		PoolName            string   `json:"pool_name"`
		Type                string   `json:"type"`
		PgAutoscaleMode     string   `json:"pg_autoscale_mode"`
		PgNum               int      `json:"pg_num"`
		Size                int      `json:"size"`
		CrushRule           string   `json:"crush_rule"`
		QuotaMaxBytes       int64    `json:"quota_max_bytes"`
		ApplicationMetadata []string `json:"application_metadata"`
	}

	var getResp GetPoolResponse
	err = json.Unmarshal(resp, &getResp)
	if err != nil {
		return nil, err
	}

	// Map response to Pool struct
	pool := &Pool{
		PoolName:            getResp.PoolName,
		Type:                getResp.Type,
		PgAutoscaleMode:     getResp.PgAutoscaleMode,
		PgNum:               getResp.PgNum,
		Size:                getResp.Size,
		RuleName:            getResp.CrushRule,
		QuotaMaxBytes:       getResp.QuotaMaxBytes,
		ApplicationMetadata: getResp.ApplicationMetadata,
	}

	return pool, nil
}

// PoolUpdate represents the fields that can be updated on a pool
type PoolUpdate struct {
	PgAutoscaleMode     string   `json:"pg_autoscale_mode,omitempty"`
	PgNum               int      `json:"pg_num,omitempty"`
	Size                int      `json:"size,omitempty"`
	QuotaMaxBytes       int64    `json:"quota_max_bytes,omitempty"`
	ApplicationMetadata []string `json:"application_metadata,omitempty"`
}

// UpdatePool updates an existing pool
func (c *Client) UpdatePool(name string, pool Pool) error {
	// Only send fields that can be updated
	update := PoolUpdate{
		PgAutoscaleMode:     pool.PgAutoscaleMode,
		PgNum:               pool.PgNum,
		Size:                pool.Size,
		QuotaMaxBytes:       pool.QuotaMaxBytes,
		ApplicationMetadata: pool.ApplicationMetadata,
	}

	rb, err := json.Marshal(update)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("PUT", fmt.Sprintf("/api/pool/%s", name), bytes.NewBuffer(rb))
	return err
}

// DeletePool deletes a pool
func (c *Client) DeletePool(name string) error {
	_, err := c.DoRequest("DELETE", fmt.Sprintf("/api/pool/%s", name), nil)
	return err
}
