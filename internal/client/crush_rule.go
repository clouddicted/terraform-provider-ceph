package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CrushRule represents a CRUSH rule
type CrushRule struct {
	Name          string `json:"name"`
	Root          string `json:"root"`
	FailureDomain string `json:"failure_domain"`
	DeviceClass   string `json:"device_class,omitempty"`
}

// CrushRuleResponse represents the API response for a CRUSH rule
type CrushRuleResponse struct {
	RuleID   int    `json:"rule_id"`
	RuleName string `json:"rule_name"`
	Type     int    `json:"type"`
	Steps    []struct {
		Op       string `json:"op"`
		Item     int    `json:"item,omitempty"`
		ItemName string `json:"item_name,omitempty"`
		Num      int    `json:"num,omitempty"`
		Type     string `json:"type,omitempty"`
	} `json:"steps"`
}

// CreateCrushRule creates a new CRUSH rule
func (c *Client) CreateCrushRule(rule CrushRule) error {
	rb, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("POST", "/api/crush_rule", bytes.NewBuffer(rb))
	return err
}

// GetCrushRule retrieves a CRUSH rule by name
func (c *Client) GetCrushRule(name string) (*CrushRuleResponse, error) {
	resp, err := c.DoRequest("GET", "/api/crush_rule", nil)
	if err != nil {
		return nil, err
	}

	var rules []CrushRuleResponse
	err = json.Unmarshal(resp, &rules)
	if err != nil {
		return nil, err
	}

	for _, r := range rules {
		if r.RuleName == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("crush rule %s not found", name)
}

// DeleteCrushRule deletes a CRUSH rule by name
func (c *Client) DeleteCrushRule(name string) error {
	_, err := c.DoRequest("DELETE", fmt.Sprintf("/api/crush_rule/%s", name), nil)
	return err
}
