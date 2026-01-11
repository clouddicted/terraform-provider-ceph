package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Capability represents a single Ceph capability entry
type Capability struct {
	Entity string `json:"entity"`
	Cap    string `json:"cap"`
}

// UserRequest represents the payload for creating/updating a Ceph user
type UserRequest struct {
	UserEntity   string       `json:"user_entity"`
	Capabilities []Capability `json:"capabilities"`
	ImportData   string       `json:"import_data,omitempty"`
}

// User represents a Ceph user (for GET responses)
type User struct {
	UserEntity   string       `json:"user_entity"`
	Capabilities []Capability `json:"capabilities"`
}

// BuildCapabilities creates the capabilities array from a list of pool names
func BuildCapabilities(pools []string) []Capability {
	caps := []Capability{
		{Entity: "mon", Cap: "allow r"},
	}
	for _, pool := range pools {
		caps = append(caps, Capability{
			Entity: "osd",
			Cap:    fmt.Sprintf("profile rbd pool=%s", pool),
		})
	}
	return caps
}

// CreateUser creates a new Ceph user
func (c *Client) CreateUser(entity string, pools []string) error {
	req := UserRequest{
		UserEntity:   entity,
		Capabilities: BuildCapabilities(pools),
	}

	rb, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("POST", "/api/cluster/user", bytes.NewBuffer(rb))
	return err
}

// UserResponse represents the API response for a user (different from create request)
type UserResponse struct {
	Entity string            `json:"entity"`
	Caps   map[string]string `json:"caps"`
	Key    string            `json:"key"`
}

// GetUser retrieves a user by entity name (e.g., client.admin)
func (c *Client) GetUser(entity string) (*UserResponse, error) {
	resp, err := c.DoRequest("GET", "/api/cluster/user", nil)
	if err != nil {
		return nil, err
	}

	var users []UserResponse
	err = json.Unmarshal(resp, &users)
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.Entity == entity {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("user %s not found", entity)
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(entity string, pools []string) error {
	req := UserRequest{
		UserEntity:   entity,
		Capabilities: BuildCapabilities(pools),
	}

	rb, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("PUT", fmt.Sprintf("/api/cluster/user/%s", url.PathEscape(entity)), bytes.NewBuffer(rb))
	return err
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(entity string) error {
	_, err := c.DoRequest("DELETE", fmt.Sprintf("/api/cluster/user/%s", url.PathEscape(entity)), nil)
	return err
}

// ExportUser retrieves the keyring/key for a user
func (c *Client) ExportUser(entity string) (string, error) {
	payload := map[string][]string{
		"entities": {entity},
	}
	rb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest("POST", "/api/cluster/user/export", bytes.NewBuffer(rb))
	if err != nil {
		return "", err
	}

	// Parse the key from keyring format
	// The response is JSON-encoded string containing keyring like:
	// "[client.testuser]\n\tkey = AQA25mJp...\n\tcaps mon = ...\n"
	keyring := strings.Trim(string(resp), "\"")
	keyring = strings.ReplaceAll(keyring, "\\n", "\n")
	keyring = strings.ReplaceAll(keyring, "\\t", "\t")

	re := regexp.MustCompile(`key\s*=\s*([A-Za-z0-9+/=]+)`)
	match := re.FindStringSubmatch(keyring)
	if len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("could not parse key from keyring")
}
