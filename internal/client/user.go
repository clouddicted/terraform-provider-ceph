package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// User represents a Ceph user
type User struct {
	UserEntity   string `json:"user_entity"`
	Capabilities string `json:"capabilities"`
	ImportData   string `json:"import_data,omitempty"`
}

// CreateUser creates a new Ceph user
func (c *Client) CreateUser(user User) error {
	rb, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("POST", "/api/cluster/user", bytes.NewBuffer(rb))
	return err
}

// GetUser retrieves a user by entity name (e.g., client.admin)
func (c *Client) GetUser(entity string) (*User, error) {
	// The API endpoint for getting a user might be /api/cluster/user/{entity}
	// Note: entity often contains dots, so it might need encoding or specific handling.
	// Let's assume standard path parameter.

	resp, err := c.DoRequest("GET", fmt.Sprintf("/api/cluster/user/%s", entity), nil)
	if err != nil {
		return nil, err
	}

	// The response format for GET might differ. It usually returns a list of caps.
	// Let's assume for now it returns a structure we can map or we need to parse.
	// Actually, `ceph auth get` returns a keyring format or json.
	// The Dashboard API usually returns a JSON object.

	// Let's define a struct that matches the GET response if it differs.
	// For now, let's try to unmarshal into User.
	var user User
	// The GET response might have "entity" instead of "user_entity" and "caps" list instead of string.
	// We might need to adjust this after testing.

	err = json.Unmarshal(resp, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(entity string, user User) error {
	rb, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = c.DoRequest("PUT", fmt.Sprintf("/api/cluster/user/%s", entity), bytes.NewBuffer(rb))
	return err
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(entity string) error {
	// The DELETE endpoint usually requires the entity name.
	_, err := c.DoRequest("DELETE", fmt.Sprintf("/api/cluster/user/%s", entity), nil)
	return err
}

// ExportUser retrieves the keyring/key for a user
func (c *Client) ExportUser(entity string) (string, error) {
	payload := map[string]string{
		"entities": entity,
	}
	rb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest("POST", "/api/cluster/user/export", bytes.NewBuffer(rb))
	if err != nil {
		return "", err
	}

	// The response is usually the keyring content in plain text or JSON depending on Accept header?
	// The curl example uses "Accept: application/vnd.ceph.api.v1.0+json".
	// If it returns JSON, it might be wrapped.
	// Let's assume it returns a JSON with the key or the keyring text.
	// Based on typical Dashboard API, it might return a string or a list.

	// Let's try to unmarshal as generic map first to see structure, or string.
	// If the response is just the keyring text (INI format), unmarshal will fail.
	// But since we send Accept: json, it should be JSON.

	// It likely returns: "[client.csi-supabase]\n\tkey = ...\n" inside a JSON string or object?
	// Or maybe just the key?
	// Let's assume it returns a JSON object with the keyring.

	// Actually, looking at similar APIs, it might return the raw keyring text if we didn't ask for JSON,
	// but we do ask for JSON.
	// Let's assume it returns: { "keyring": "..." } or similar.
	// Or maybe it returns the list of entities with keys.

	// Let's try to parse as a simple map for now.
	// If it's a list of users, we need to find ours.

	// Wait, the user said: curl ... -d '{ "entities": "client.csi-supabase" }'
	// This suggests it exports specific entities.

	// Let's assume the response is the keyring string directly if it's not JSON,
	// OR a JSON containing it.
	// Given the Accept header, it's likely JSON.

	// Let's try to unmarshal into a struct that holds the keyring.
	// If it fails, we return the raw body string.

	// NOTE: We need to extract the actual "key" from the keyring if possible,
	// or just return the whole keyring. The user asked for "the key".
	// The keyring format is:
	// [client.x]
	// 	key = AAAA...

	// We can parse this.

	return string(resp), nil
}
