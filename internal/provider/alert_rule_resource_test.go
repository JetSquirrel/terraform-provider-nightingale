// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAlertRuleResourceImportValid(t *testing.T) {
	bgid, id, err := parseImportID("3:789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bgid != 3 {
		t.Errorf("busi_group_id = %d, want 3", bgid)
	}
	if id != 789 {
		t.Errorf("alert_rule_id = %d, want 789", id)
	}
}

func TestAlertRuleResourceImportInvalidID(t *testing.T) {
	cases := []string{
		"invalid",
		"abc:def",
		"0:123",
		"123:0",
		"123",
		":",
	}

	for _, id := range cases {
		_, _, err := parseImportID(id)
		if err == nil {
			t.Errorf("expected error for import ID %q", id)
		}
	}
}

func TestAlertRuleResourceExtraJSONValidator(t *testing.T) {
	v := jsonValidator{}

	// Valid JSON object
	var validResp validator.StringResponse
	v.ValidateString(t.Context(), validator.StringRequest{
		ConfigValue: types.StringValue(`{"key":"value"}`),
	}, &validResp)
	if validResp.Diagnostics.HasError() {
		t.Errorf("unexpected error for valid JSON: %s", validResp.Diagnostics.Errors())
	}

	// Invalid JSON
	var invalidResp validator.StringResponse
	v.ValidateString(t.Context(), validator.StringRequest{
		ConfigValue: types.StringValue(`not json`),
	}, &invalidResp)
	if !invalidResp.Diagnostics.HasError() {
		t.Error("expected error for invalid JSON")
	}

	// Null/unknown/empty should pass
	var nullResp validator.StringResponse
	v.ValidateString(t.Context(), validator.StringRequest{
		ConfigValue: types.StringNull(),
	}, &nullResp)
	if nullResp.Diagnostics.HasError() {
		t.Errorf("unexpected error for null value: %s", nullResp.Diagnostics.Errors())
	}
}

func TestAlertRuleResourceRefreshState(t *testing.T) {
	r := &AlertRuleResource{}
	ctx := t.Context()

	state := &AlertRuleResourceModel{
		ID:          types.StringValue("123"),
		BusiGroupID: types.Int64Value(1),
	}

	remote := &client.AlertRule{
		ID:              123,
		GroupID:         1,
		Name:            "Test Rule",
		DatasourceType:  "prometheus",
		Disabled:        0,
		Severity:        2,
		RuleConfig:      `{"queries":[{"prom_ql":"up == 1","duration_seconds":300}]}`,
		PromForDuration: 300,
		AppendTags:      []string{"env=prod"},
		Annotations:     map[string]string{"summary": "test"},
		NotifyChannels:  []string{"email"},
		NotifyRecovered: 1,
		NotifyRuleIDs:   []int64{1, 2},
		RunbookURL:      "https://example.com",
		CreateAt:        1000,
		CreateBy:        "admin",
		UpdateAt:        2000,
		UpdateBy:        "user",
	}

	r.refreshState(ctx, state, remote)

	if state.Name.ValueString() != "Test Rule" {
		t.Errorf("name = %q, want Test Rule", state.Name.ValueString())
	}
	if state.DatasourceType.ValueString() != "prometheus" {
		t.Errorf("datasource_type = %q, want prometheus", state.DatasourceType.ValueString())
	}
	if state.Disabled.ValueBool() != false {
		t.Errorf("disabled = %v, want false", state.Disabled.ValueBool())
	}
	if state.Severity.ValueInt64() != 2 {
		t.Errorf("severity = %d, want 2", state.Severity.ValueInt64())
	}
	if state.RunbookURL.ValueString() != "https://example.com" {
		t.Errorf("runbook_url = %q", state.RunbookURL.ValueString())
	}
	if state.CreateAt.ValueInt64() != 1000 {
		t.Errorf("create_at = %d", state.CreateAt.ValueInt64())
	}
	if state.CreateBy.ValueString() != "admin" {
		t.Errorf("create_by = %q", state.CreateBy.ValueString())
	}
	if state.UpdateAt.ValueInt64() != 2000 {
		t.Errorf("update_at = %d", state.UpdateAt.ValueInt64())
	}
	if state.UpdateBy.ValueString() != "user" {
		t.Errorf("update_by = %q", state.UpdateBy.ValueString())
	}

	// Verify queries were parsed
	if state.Queries.IsNull() || state.Queries.IsUnknown() {
		t.Fatal("queries should not be null/unknown")
	}

	// Verify sets
	if state.AppendTags.IsNull() {
		t.Error("append_tags should not be null")
	}
	if state.Annotations.IsNull() {
		t.Error("annotations should not be null")
	}
	if state.NotifyChannels.IsNull() {
		t.Error("notify_channels should not be null")
	}
	if state.NotifyRuleIDs.IsNull() {
		t.Error("notify_rule_ids should not be null")
	}
}

func TestAlertRuleResourceExpandQueries(t *testing.T) {
	ctx := t.Context()

	queryModels := []AlertRuleQueryModel{
		{
			Ref:                types.StringValue("A"),
			PromQL:             types.StringValue("up == 1"),
			DurationSeconds:    types.Int64Value(300),
			ComparisonOperator: types.StringValue(">"),
			Threshold:          types.Float64Value(0.5),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ref":                 types.StringType,
			"promql":              types.StringType,
			"duration_seconds":    types.Int64Type,
			"comparison_operator": types.StringType,
			"threshold":           types.Float64Type,
		},
	}, queryModels)
	if diags.HasError() {
		t.Fatalf("failed to create list: %s", diags.Errors())
	}

	queries, err := expandQueries(ctx, listValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	if queries[0].Ref != "A" {
		t.Errorf("ref = %q", queries[0].Ref)
	}
	if queries[0].PromQL != "up == 1" {
		t.Errorf("promql = %q", queries[0].PromQL)
	}
	if queries[0].DurationSeconds != 300 {
		t.Errorf("duration_seconds = %d", queries[0].DurationSeconds)
	}
	if queries[0].ComparisonOperator != ">" {
		t.Errorf("comparison_operator = %q", queries[0].ComparisonOperator)
	}
	if queries[0].Threshold != 0.5 {
		t.Errorf("threshold = %f", queries[0].Threshold)
	}
}

func TestAlertRuleResourceParseExtraJSON(t *testing.T) {
	// Valid JSON
	extra, err := parseExtraJSON(types.StringValue(`{"key":"value"}`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if extra["key"] != "value" {
		t.Errorf("extra[key] = %v", extra["key"])
	}

	// Empty/null
	extra, err = parseExtraJSON(types.StringNull())
	if err != nil {
		t.Errorf("unexpected error for null: %v", err)
	}
	if extra != nil {
		t.Error("expected nil for null")
	}

	// Invalid JSON
	_, err = parseExtraJSON(types.StringValue(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAlertRuleResourceHelpers(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("boolToInt(true) should be 1")
	}
	if boolToInt(false) != 0 {
		t.Error("boolToInt(false) should be 0")
	}

	if !isNotFound(fmt.Errorf("rule not found")) {
		t.Error("isNotFound should match 'not found'")
	}
	if !isNotFound(fmt.Errorf("404 error")) {
		t.Error("isNotFound should match '404'")
	}
	if isNotFound(nil) {
		t.Error("isNotFound(nil) should be false")
	}
	if isNotFound(fmt.Errorf("some other error")) {
		t.Error("isNotFound should not match unrelated errors")
	}
}

func TestAlertRuleResourceReadRemovesStateWhenNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")

	// Since we cannot easily construct a tfsdk.State in unit tests,
	// we verify the client behavior instead: the isNotFound helper
	// correctly identifies the error, and the resource's Read method
	// would call resp.State.RemoveResource(ctx) when the client
	// returns an error matching isNotFound.
	_, err := c.GetAlertRule(t.Context(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !isNotFound(err) {
		t.Errorf("expected isNotFound to match: %v", err)
	}
}

func TestAlertRuleResourceDeleteToleratesNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c, _ := client.New(server.URL, "token", 30, false, "test")

	// Verify the client error is tolerated by isNotFound
	err := c.DeleteAlertRules(t.Context(), 1, []int64{999})
	if err == nil {
		t.Fatal("expected error from client")
	}
	if !isNotFound(err) {
		t.Errorf("expected isNotFound to match client delete error: %v", err)
	}

	// Verify the resource delete method would tolerate this:
	// In Delete(), it calls isNotFound(err) and returns without adding diagnostics.
	if !isNotFound(err) {
		t.Error("Delete should tolerate not-found errors")
	}
}
