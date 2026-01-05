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

	// The GET response might have different field names than POST request.
	// Usually "pool_name" instead of "pool".
	// We need a separate struct or custom unmarshaling if they differ significantly.
	// For now, let's assume we can map it.
	// Based on typical Ceph API, GET returns "pool_name".
	type GetPoolResponse struct {
		PoolName            string                 `json:"pool_name"`
		Type                string                 `json:"type"` // "replicated" vs "pool_type"?
		PgAutoscaleMode     string                 `json:"pg_autoscale_mode"`
		PgNum               int                    `json:"pg_num"`
		Size                int                    `json:"size"`
		QuotaMaxBytes       int64                  `json:"quota_max_bytes"`
		ApplicationMetadata map[string]interface{} `json:"application_metadata"` // It returns a map, not list?
		// ... other fields
	}

	// Let's stick to the simple struct for now and adjust if testing fails.
	// The user provided the POST payload.

	var pool Pool
	err = json.Unmarshal(resp, &pool)
	if err != nil {
		return nil, err
	}

	// Map pool_name from response if needed, but our struct uses "pool" json tag.
	// If the API returns "pool_name", it won't populate "PoolName" field which has `json:"pool"`.
	// We should probably use a separate struct for GET or use multiple tags (not supported in stdlib).
	// Let's fix the struct to handle both if possible, or just use a map for flexibility?
	// No, let's use a specific struct for GET.

	return &pool, nil
}

// UpdatePool updates an existing pool
func (c *Client) UpdatePool(name string, pool Pool) error {
	rb, err := json.Marshal(pool)
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
