package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client holds the connection details
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

// AuthResponse represents the response from the login endpoint
type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// NewClient creates a new Ceph Dashboard API client
func NewClient(host, username, password string, insecure bool) (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	c := &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second, Transport: tr},
		HostURL:    host,
	}

	// Authenticate immediately
	if username != "" && password != "" {
		err := c.SignIn(username, password)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// SignIn authenticates against the Ceph Dashboard and stores the token
func (c *Client) SignIn(username, password string) error {
	authPayload := map[string]string{
		"username": username,
		"password": password,
	}
	rb, err := json.Marshal(authPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/auth", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d, error: %s", res.StatusCode, "authentication failed")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	ar := AuthResponse{}
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return err
	}

	c.Token = ar.Token
	return nil
}

// DoRequest performs the HTTP request with the authenticated token
func (c *Client) DoRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	return c.DoRequestWithHeaders(method, endpoint, body, nil)
}

// DoRequestWithHeaders performs the HTTP request with custom headers
func (c *Client) DoRequestWithHeaders(method, endpoint string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.HostURL, endpoint), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// Default Accept header, can be overridden
	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, string(respBody))
	}

	return respBody, nil
}
