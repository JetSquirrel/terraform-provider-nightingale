// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Envelope is the standard Nightingale JSON response envelope.
type Envelope struct {
	Dat json.RawMessage `json:"dat"`
	Err string          `json:"err"`
}

// Client is a typed HTTP client for the Nightingale page-operation API.
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
	userAgent  string
}

// New creates a new Nightingale client.
func New(endpoint, token string, timeoutSeconds int, insecureSkipTLSVerify bool, userAgent string) (*Client, error) {
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return nil, fmt.Errorf("endpoint must have http or https scheme: %s", endpoint)
	}
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	transport := &http.Transport{}
	if insecureSkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		endpoint: endpoint,
		token:    token,
		httpClient: &http.Client{
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
			Transport: transport,
		},
		userAgent: userAgent,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, uri string, body interface{}) (*Envelope, error) {
	var bodyReader io.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bs)
	}

	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	u.Path = path.Join(u.Path, uri)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-User-Token", c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview := string(respBody)
		if len(preview) > 512 {
			preview = preview[:512] + "..."
		}
		return nil, fmt.Errorf("unexpected status code: method=%s path=%s status=%d body=%s", method, uri, resp.StatusCode, preview)
	}

	var env Envelope
	if len(respBody) == 0 || string(respBody) == "null" {
		env = Envelope{Dat: json.RawMessage("null"), Err: ""}
	} else if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if env.Err != "" {
		return nil, fmt.Errorf("api error: %s", env.Err)
	}

	return &env, nil
}
