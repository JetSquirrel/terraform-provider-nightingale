// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// AlertRuleQuery represents a single query in an alert rule.
type AlertRuleQuery struct {
	Ref                string  `json:"ref,omitempty"`
	PromQL             string  `json:"prom_ql"`
	DurationSeconds    int64   `json:"duration_seconds,omitempty"`
	ComparisonOperator string  `json:"comparison_operator,omitempty"`
	Threshold          float64 `json:"threshold,omitempty"`
}

// FlexibleRuleConfig handles rule_config that can be either a string or an object.
type FlexibleRuleConfig string

func (f *FlexibleRuleConfig) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexibleRuleConfig(s)
		return nil
	}
	*f = FlexibleRuleConfig(string(data))
	return nil
}

func (f FlexibleRuleConfig) MarshalJSON() ([]byte, error) {
	if f == "" {
		return []byte(`""`), nil
	}
	return []byte(f), nil
}

// AlertRule represents a Nightingale alert rule.
type AlertRule struct {
	ID              int64              `json:"id"`
	GroupID         int64              `json:"group_id"`
	Name            string             `json:"name"`
	DatasourceType  string             `json:"cate"`
	DatasourceIDs   []int64            `json:"datasource_ids"`
	Disabled        int                `json:"disabled"`
	Severity        int64              `json:"severity"`
	RuleConfig      FlexibleRuleConfig `json:"rule_config"`
	PromForDuration int64              `json:"prom_for_duration"`
	AppendTags      []string           `json:"append_tags"`
	Annotations     map[string]string  `json:"annotations"`
	NotifyChannels  []string           `json:"notify_channels"`
	NotifyRecovered int                `json:"notify_recovered"`
	NotifyRuleIDs   []int64            `json:"notify_rule_ids"`
	RunbookURL      string             `json:"runbook_url"`
	CreateAt        int64              `json:"create_at"`
	CreateBy        string             `json:"create_by"`
	UpdateAt        int64              `json:"update_at"`
	UpdateBy        string             `json:"update_by"`
}

// RuleConfigPayload is the structure stored in the RuleConfig JSON field.
type RuleConfigPayload struct {
	Queries []AlertRuleQuery `json:"queries"`
}

func BuildRuleConfig(queries []AlertRuleQuery) (string, error) {
	if len(queries) == 0 {
		return "", nil
	}
	cfg := RuleConfigPayload{Queries: queries}
	bs, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func ParseRuleConfig(ruleConfig string) ([]AlertRuleQuery, error) {
	if ruleConfig == "" {
		return nil, nil
	}
	var cfg RuleConfigPayload
	if err := json.Unmarshal([]byte(ruleConfig), &cfg); err != nil {
		return nil, err
	}
	return cfg.Queries, nil
}

func toPayload(rule *AlertRule, extra map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"name":              rule.Name,
		"cate":              rule.DatasourceType,
		"datasource_ids":    rule.DatasourceIDs,
		"disabled":          rule.Disabled,
		"severity":          rule.Severity,
		"rule_config":       string(rule.RuleConfig),
		"prom_for_duration": rule.PromForDuration,
		"append_tags":       rule.AppendTags,
		"annotations":       rule.Annotations,
		"notify_channels":   rule.NotifyChannels,
		"notify_recovered":  rule.NotifyRecovered,
		"notify_rule_ids":   rule.NotifyRuleIDs,
		"runbook_url":       rule.RunbookURL,
	}

	for k, v := range extra {
		payload[k] = v
	}

	return payload, nil
}

// ListAlertRules lists all alert rules in a business group.
func (c *Client) ListAlertRules(ctx context.Context, groupID int64) ([]AlertRule, error) {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-rules", groupID)

	env, err := c.doRequest(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}

	var rules []AlertRule
	if err := json.Unmarshal(env.Dat, &rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert rules: %w", err)
	}

	return rules, nil
}

// CreateAlertRule creates a new alert rule in the specified business group.
func (c *Client) CreateAlertRule(ctx context.Context, groupID int64, rule *AlertRule, extra map[string]interface{}) (*AlertRule, error) {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-rules", groupID)

	payload, err := toPayload(rule, extra)
	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(ctx, "POST", uri, []map[string]interface{}{payload})
	if err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	rules, err := c.ListAlertRules(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created alert rule: %w", err)
	}

	var latestRule *AlertRule
	for i := range rules {
		if rules[i].Name == rule.Name {
			if latestRule == nil || rules[i].CreateAt > latestRule.CreateAt {
				latestRule = &rules[i]
			}
		}
	}

	if latestRule == nil {
		return nil, fmt.Errorf("created alert rule not found in list")
	}

	return latestRule, nil
}

// GetAlertRule retrieves a single alert rule by ID.
func (c *Client) GetAlertRule(ctx context.Context, id int64) (*AlertRule, error) {
	uri := fmt.Sprintf("/api/n9e/alert-rule/%d", id)

	env, err := c.doRequest(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	var rule AlertRule
	if err := json.Unmarshal(env.Dat, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert rule: %w", err)
	}

	return &rule, nil
}

// UpdateAlertRule updates an existing alert rule.
func (c *Client) UpdateAlertRule(ctx context.Context, groupID, id int64, rule *AlertRule, extra map[string]interface{}) (*AlertRule, error) {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-rule/%d", groupID, id)

	payload, err := toPayload(rule, extra)
	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(ctx, "PUT", uri, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update alert rule: %w", err)
	}

	return c.GetAlertRule(ctx, id)
}

// DeleteAlertRules deletes alert rules by IDs in a business group.
func (c *Client) DeleteAlertRules(ctx context.Context, groupID int64, ids []int64) error {
	uri := fmt.Sprintf("/api/n9e/busi-group/%d/alert-rules", groupID)

	payload := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.doRequest(ctx, "DELETE", uri, payload)
	if err != nil {
		return fmt.Errorf("failed to delete alert rules: %w", err)
	}

	return nil
}
