// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNotifyRuleResourceRefreshState(t *testing.T) {
	r := &NotifyRuleResource{}
	ctx := context.Background()

	state := &NotifyRuleResourceModel{ID: types.StringValue("11")}
	remote := &client.NotifyRule{
		ID:           11,
		Name:         "Test Notify",
		Enable:       1,
		UserGroupIds: []int64{1, 2},
		NotifyConfigs: []client.NotifyConfig{
			{ChannelID: 1, TemplateID: 2, Params: map[string]interface{}{"key": "val"}},
		},
		CreateAt: 1000,
		CreateBy: "admin",
		UpdateAt: 2000,
		UpdateBy: "user",
	}

	r.refreshState(ctx, state, remote)

	if state.Name.ValueString() != "Test Notify" {
		t.Errorf("name = %q", state.Name.ValueString())
	}
	if !state.Enable.ValueBool() {
		t.Error("enable should be true")
	}
	if state.CreateAt.ValueInt64() != 1000 {
		t.Errorf("create_at = %d", state.CreateAt.ValueInt64())
	}
	if state.UserGroupIds.IsNull() {
		t.Error("user_group_ids should not be null")
	}
	if state.NotifyConfigs.IsNull() {
		t.Error("notify_configs should not be null")
	}
}

func TestNotifyRuleResourceExpandConfigs(t *testing.T) {
	ctx := context.Background()
	configModels := []NotifyConfigModel{
		{
			ChannelID:  types.Int64Value(1),
			TemplateID: types.Int64Value(2),
			Params:     types.MapNull(types.StringType),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"channel_id":  types.Int64Type,
			"template_id": types.Int64Type,
			"params":      types.MapType{ElemType: types.StringType},
		},
	}, configModels)
	if diags.HasError() {
		t.Fatalf("failed to create list: %s", diags.Errors())
	}

	configs, err := expandNotifyConfigs(ctx, listValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].ChannelID != 1 {
		t.Errorf("channel_id = %d", configs[0].ChannelID)
	}
	if configs[0].TemplateID != 2 {
		t.Errorf("template_id = %d", configs[0].TemplateID)
	}
}

func TestNotifyRuleResourceReadNotFound(t *testing.T) {
	server := newMockServer(404, "not found")
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")
	_, err := c.GetNotifyRule(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	if !isNotFound(err) {
		t.Errorf("expected not found: %v", err)
	}
}

func TestNotifyRuleResourceDeleteToleratesNotFound(t *testing.T) {
	server := newMockServer(404, "not found")
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")
	err := c.DeleteNotifyRules(context.Background(), []int64{999})
	if err == nil {
		t.Fatal("expected error from client")
	}
	if !isNotFound(err) {
		t.Errorf("expected isNotFound to match: %v", err)
	}
}

func newMockServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
}
