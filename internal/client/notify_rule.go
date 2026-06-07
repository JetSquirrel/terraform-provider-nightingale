package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// NotifyConfig represents a single notification configuration within a notify rule.
type NotifyConfig struct {
	ChannelID  int64                  `json:"channel_id"`
	TemplateID int64                  `json:"template_id,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// NotifyRule represents a Nightingale notification rule.
type NotifyRule struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Enable        int            `json:"enable"`
	UserGroupIds  []int64        `json:"user_group_ids"`
	NotifyConfigs []NotifyConfig `json:"notify_configs"`
	CreateAt      int64          `json:"create_at"`
	CreateBy      string         `json:"create_by"`
	UpdateAt      int64          `json:"update_at"`
	UpdateBy      string         `json:"update_by"`
}

// CreateNotifyRule creates a new notification rule.
func (c *Client) CreateNotifyRule(ctx context.Context, rule *NotifyRule) (*NotifyRule, error) {
	uri := "/api/n9e/notify-rules"

	env, err := c.doRequest(ctx, "POST", uri, []*NotifyRule{rule})
	if err != nil {
		return nil, fmt.Errorf("failed to create notify rule: %w", err)
	}

	var created []*NotifyRule
	if err := json.Unmarshal(env.Dat, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created notify rule: %w", err)
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("create notify rule returned empty list")
	}

	return created[0], nil
}

// GetNotifyRule retrieves a single notification rule by ID.
func (c *Client) GetNotifyRule(ctx context.Context, id int64) (*NotifyRule, error) {
	uri := fmt.Sprintf("/api/n9e/notify-rule/%d", id)

	env, err := c.doRequest(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get notify rule: %w", err)
	}

	var rule NotifyRule
	if err := json.Unmarshal(env.Dat, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notify rule: %w", err)
	}

	return &rule, nil
}

// UpdateNotifyRule updates an existing notification rule.
func (c *Client) UpdateNotifyRule(ctx context.Context, id int64, rule *NotifyRule) (*NotifyRule, error) {
	uri := fmt.Sprintf("/api/n9e/notify-rule/%d", id)

	env, err := c.doRequest(ctx, "PUT", uri, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update notify rule: %w", err)
	}

	var updated NotifyRule
	if err := json.Unmarshal(env.Dat, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated notify rule: %w", err)
	}

	return &updated, nil
}

// DeleteNotifyRules deletes notification rules by IDs.
func (c *Client) DeleteNotifyRules(ctx context.Context, ids []int64) error {
	uri := "/api/n9e/notify-rules"

	payload := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.doRequest(ctx, "DELETE", uri, payload)
	if err != nil {
		return fmt.Errorf("failed to delete notify rules: %w", err)
	}

	return nil
}
