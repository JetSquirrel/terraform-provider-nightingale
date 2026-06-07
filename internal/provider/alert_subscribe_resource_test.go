// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAlertSubscribeResourceRefreshState(t *testing.T) {
	r := &AlertSubscribeResource{}
	ctx := context.Background()

	state := &AlertSubscribeResourceModel{
		ID:          types.StringValue("22"),
		BusiGroupID: types.Int64Value(2),
	}
	remote := &client.AlertSubscribe{
		ID:            22,
		GroupId:       2,
		Name:          "Test Sub",
		Disabled:      0,
		DatasourceIds: []int64{1},
		RuleIds:       []int64{10, 11},
		Severities:    []int64{1, 2},
		Tags:          "env=prod",
		BusiGroups:    "ops",
		UserGroupIds:  []int64{5},
		NotifyRuleIds: []int64{3},
		NotifyVersion: 1,
		CreateAt:      1000,
		CreateBy:      "admin",
		UpdateAt:      2000,
		UpdateBy:      "user",
	}

	r.refreshState(ctx, state, remote)

	if state.Name.ValueString() != "Test Sub" {
		t.Errorf("name = %q", state.Name.ValueString())
	}
	if state.Disabled.ValueBool() {
		t.Error("disabled should be false")
	}
	if state.Tags.ValueString() != "env=prod" {
		t.Errorf("tags = %q", state.Tags.ValueString())
	}
	if state.BusiGroups.ValueString() != "ops" {
		t.Errorf("busi_groups = %q", state.BusiGroups.ValueString())
	}
	if state.NotifyVersion.ValueInt64() != 1 {
		t.Errorf("notify_version = %d", state.NotifyVersion.ValueInt64())
	}
	if state.DatasourceIds.IsNull() {
		t.Error("datasource_ids should not be null")
	}
	if state.RuleIds.IsNull() {
		t.Error("rule_ids should not be null")
	}
	if state.Severities.IsNull() {
		t.Error("severities should not be null")
	}
	if state.UserGroupIds.IsNull() {
		t.Error("user_group_ids should not be null")
	}
	if state.NotifyRuleIds.IsNull() {
		t.Error("notify_rule_ids should not be null")
	}
}

func TestAlertSubscribeResourceToAPI(t *testing.T) {
	r := &AlertSubscribeResource{}
	state := &AlertSubscribeResourceModel{
		BusiGroupID:   types.Int64Value(3),
		Name:          types.StringValue("Sub Name"),
		Disabled:      types.BoolValue(true),
		Tags:          types.StringValue("team=ops"),
		NotifyVersion: types.Int64Value(1),
	}

	api := r.toAPI(state)
	if api.GroupId != 3 {
		t.Errorf("group_id = %d", api.GroupId)
	}
	if api.Name != "Sub Name" {
		t.Errorf("name = %q", api.Name)
	}
	if api.Disabled != 1 {
		t.Errorf("disabled = %d", api.Disabled)
	}
	if api.Tags != "team=ops" {
		t.Errorf("tags = %q", api.Tags)
	}
	if api.NotifyVersion != 1 {
		t.Errorf("notify_version = %d", api.NotifyVersion)
	}
}

func TestAlertSubscribeResourceReadNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")
	_, err := c.GetAlertSubscribe(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	if !isNotFound(err) {
		t.Errorf("expected not found: %v", err)
	}
}

func TestAlertSubscribeResourceDeleteToleratesNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")
	err := c.DeleteAlertSubscribes(context.Background(), 1, []int64{999})
	if err == nil {
		t.Fatal("expected error from client")
	}
	if !isNotFound(err) {
		t.Errorf("expected isNotFound to match: %v", err)
	}
}
