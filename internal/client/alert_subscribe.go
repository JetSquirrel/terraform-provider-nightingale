// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// AlertSubscribe represents a Nightingale alert subscription rule.
type AlertSubscribe struct {
	ID            int64   `json:"id"`
	GroupId       int64   `json:"group_id"`
	Name          string  `json:"name"`
	Disabled      int     `json:"disabled"`
	DatasourceIds []int64 `json:"datasource_ids"`
	RuleIds       []int64 `json:"rule_ids"`
	Severities    []int64 `json:"severities"`
	Tags          string  `json:"tags"`
	BusiGroups    string  `json:"busi_groups"`
	UserGroupIds  []int64 `json:"user_group_ids"`
	NotifyRuleIds []int64 `json:"notify_rule_ids"`
	NotifyVersion int     `json:"notify_version"`
	CreateAt      int64   `json:"create_at"`
	CreateBy      string  `json:"create_by"`
	UpdateAt      int64   `json:"update_at"`
	UpdateBy      string  `json:"update_by"`
}

// CreateAlertSubscribe creates a new alert subscription in the specified business group.
func (c *Client) CreateAlertSubscribe(ctx context.Context, groupID int64, sub *AlertSubscribe) (*AlertSubscribe, error) {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-subscribes", groupID)

	env, err := c.doRequest(ctx, "POST", uri, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert subscribe: %w", err)
	}

	var created AlertSubscribe
	if err := json.Unmarshal(env.Dat, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created alert subscribe: %w", err)
	}

	return &created, nil
}

// GetAlertSubscribe retrieves a single alert subscription by ID.
func (c *Client) GetAlertSubscribe(ctx context.Context, id int64) (*AlertSubscribe, error) {
	uri := fmt.Sprintf("/api/n9e/alert-subscribe/%d", id)

	env, err := c.doRequest(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert subscribe: %w", err)
	}

	var sub AlertSubscribe
	if err := json.Unmarshal(env.Dat, &sub); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert subscribe: %w", err)
	}

	return &sub, nil
}

// UpdateAlertSubscribe updates an existing alert subscription.
// Nightingale uses a batch endpoint; we send a single-element list.
func (c *Client) UpdateAlertSubscribe(ctx context.Context, groupID int64, sub *AlertSubscribe) (*AlertSubscribe, error) {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-subscribes", groupID)

	env, err := c.doRequest(ctx, "PUT", uri, []*AlertSubscribe{sub})
	if err != nil {
		return nil, fmt.Errorf("failed to update alert subscribe: %w", err)
	}

	// The API may return null on success; treat as success
	if string(env.Dat) == "null" || len(env.Dat) == 0 {
		return sub, nil
	}

	var updated AlertSubscribe
	if err := json.Unmarshal(env.Dat, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated alert subscribe: %w", err)
	}

	return &updated, nil
}

// DeleteAlertSubscribes deletes alert subscriptions by IDs in a business group.
func (c *Client) DeleteAlertSubscribes(ctx context.Context, groupID int64, ids []int64) error {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-subscribes", groupID)

	payload := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.doRequest(ctx, "DELETE", uri, payload)
	if err != nil {
		return fmt.Errorf("failed to delete alert subscribes: %w", err)
	}

	return nil
}
