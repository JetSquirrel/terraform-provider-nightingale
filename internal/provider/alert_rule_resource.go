// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AlertRuleResource{}
var _ resource.ResourceWithImportState = &AlertRuleResource{}

func NewAlertRuleResource() resource.Resource {
	return &AlertRuleResource{}
}

type AlertRuleResource struct {
	client *client.Client
}

type AlertRuleQueryModel struct {
	Ref                types.String  `tfsdk:"ref"`
	PromQL             types.String  `tfsdk:"promql"`
	DurationSeconds    types.Int64   `tfsdk:"duration_seconds"`
	ComparisonOperator types.String  `tfsdk:"comparison_operator"`
	Threshold          types.Float64 `tfsdk:"threshold"`
}

type AlertRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	BusiGroupID     types.Int64  `tfsdk:"busi_group_id"`
	Name            types.String `tfsdk:"name"`
	DatasourceType  types.String `tfsdk:"datasource_type"`
	Queries         types.List   `tfsdk:"queries"`
	Disabled        types.Bool   `tfsdk:"disabled"`
	Severity        types.Int64  `tfsdk:"severity"`
	DatasourceIDs   types.Set    `tfsdk:"datasource_ids"`
	AppendTags      types.Set    `tfsdk:"append_tags"`
	Annotations     types.Map    `tfsdk:"annotations"`
	NotifyRuleIDs   types.Set    `tfsdk:"notify_rule_ids"`
	NotifyRecovered types.Bool   `tfsdk:"notify_recovered"`
	NotifyChannels  types.Set    `tfsdk:"notify_channels"`
	RunbookURL      types.String `tfsdk:"runbook_url"`
	ExtraJSON       types.String `tfsdk:"extra_json"`
	CreateAt        types.Int64  `tfsdk:"create_at"`
	CreateBy        types.String `tfsdk:"create_by"`
	UpdateAt        types.Int64  `tfsdk:"update_at"`
	UpdateBy        types.String `tfsdk:"update_by"`
}

func (r *AlertRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rule"
}

func (r *AlertRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightingale alert rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Nightingale alert rule ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"busi_group_id": schema.Int64Attribute{
				MarkdownDescription: "Nightingale business group ID that owns the alert rule.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Alert rule name.",
				Required:            true,
			},
			"datasource_type": schema.StringAttribute{
				MarkdownDescription: "Nightingale datasource type, for example prometheus.",
				Required:            true,
			},
			"queries": schema.ListNestedAttribute{
				MarkdownDescription: "Alert query definitions.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ref": schema.StringAttribute{
							MarkdownDescription: "Query ref, for example A.",
							Optional:            true,
						},
						"promql": schema.StringAttribute{
							MarkdownDescription: "PromQL expression.",
							Required:            true,
						},
						"duration_seconds": schema.Int64Attribute{
							MarkdownDescription: "Evaluation duration/for time.",
							Optional:            true,
						},
						"comparison_operator": schema.StringAttribute{
							MarkdownDescription: "Operator if Nightingale version uses threshold conditions outside PromQL.",
							Optional:            true,
						},
						"threshold": schema.Float64Attribute{
							MarkdownDescription: "Threshold if Nightingale version uses threshold conditions outside PromQL.",
							Optional:            true,
						},
					},
				},
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the alert rule is disabled. Default is false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"severity": schema.Int64Attribute{
				MarkdownDescription: "Nightingale alert severity.",
				Optional:            true,
			},
			"datasource_ids": schema.SetAttribute{
				MarkdownDescription: "Datasource IDs used by the rule.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"append_tags": schema.SetAttribute{
				MarkdownDescription: "Tags appended to generated alert events.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"annotations": schema.MapAttribute{
				MarkdownDescription: "User-facing annotations/metadata.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"notify_rule_ids": schema.SetAttribute{
				MarkdownDescription: "Notification rule IDs.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"notify_recovered": schema.BoolAttribute{
				MarkdownDescription: "Whether to notify on recovery.",
				Optional:            true,
			},
			"notify_channels": schema.SetAttribute{
				MarkdownDescription: "Notification channels if supported by the target Nightingale version.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"runbook_url": schema.StringAttribute{
				MarkdownDescription: "Optional runbook URL if supported/mapped through annotations.",
				Optional:            true,
			},
			"extra_json": schema.StringAttribute{
				MarkdownDescription: "JSON object merged into API payload for Nightingale-version-specific fields.",
				Optional:            true,
				Validators: []validator.String{
					jsonValidator{},
				},
			},
			"create_at": schema.Int64Attribute{
				MarkdownDescription: "Remote creation timestamp.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"create_by": schema.StringAttribute{
				MarkdownDescription: "Remote creator.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"update_at": schema.Int64Attribute{
				MarkdownDescription: "Remote update timestamp.",
				Computed:            true,
			},
			"update_by": schema.StringAttribute{
				MarkdownDescription: "Remote updater.",
				Computed:            true,
			},
		},
	}
}

func (r *AlertRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *AlertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extra, err := parseExtraJSON(plan.ExtraJSON)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	queries, err := expandQueries(ctx, plan.Queries)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	ruleConfig, err := client.BuildRuleConfig(queries)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to build rule config: %s", err))
		return
	}

	var promForDuration int64
	if len(queries) > 0 && queries[0].DurationSeconds > 0 {
		promForDuration = queries[0].DurationSeconds
	}

	apiRule := &client.AlertRule{
		GroupID:         plan.BusiGroupID.ValueInt64(),
		Name:            plan.Name.ValueString(),
		DatasourceType:  plan.DatasourceType.ValueString(),
		DatasourceIDs:   setToInt64Slice(plan.DatasourceIDs),
		Disabled:        boolToInt(plan.Disabled.ValueBool()),
		Severity:        plan.Severity.ValueInt64(),
		RuleConfig:      ruleConfig,
		PromForDuration: promForDuration,
		AppendTags:      setToStringSlice(plan.AppendTags),
		Annotations:     mapToStringMap(plan.Annotations),
		NotifyChannels:  setToStringSlice(plan.NotifyChannels),
		NotifyRecovered: boolToInt(plan.NotifyRecovered.ValueBool()),
		NotifyRuleIDs:   setToInt64Slice(plan.NotifyRuleIDs),
		RunbookURL:      plan.RunbookURL.ValueString(),
	}

	created, err := r.client.CreateAlertRule(ctx, plan.BusiGroupID.ValueInt64(), apiRule, extra)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert rule, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(created.ID, 10))
	r.refreshState(ctx, &plan, created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert rule ID: %s", err))
		return
	}

	remote, err := r.client.GetAlertRule(ctx, id)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert rule, got error: %s", err))
		return
	}

	r.refreshState(ctx, &state, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert rule ID: %s", err))
		return
	}

	extra, err := parseExtraJSON(plan.ExtraJSON)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	queries, err := expandQueries(ctx, plan.Queries)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	ruleConfig, err := client.BuildRuleConfig(queries)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to build rule config: %s", err))
		return
	}

	var promForDuration int64
	if len(queries) > 0 && queries[0].DurationSeconds > 0 {
		promForDuration = queries[0].DurationSeconds
	}

	apiRule := &client.AlertRule{
		ID:              id,
		GroupID:         plan.BusiGroupID.ValueInt64(),
		Name:            plan.Name.ValueString(),
		DatasourceType:  plan.DatasourceType.ValueString(),
		DatasourceIDs:   setToInt64Slice(plan.DatasourceIDs),
		Disabled:        boolToInt(plan.Disabled.ValueBool()),
		Severity:        plan.Severity.ValueInt64(),
		RuleConfig:      ruleConfig,
		PromForDuration: promForDuration,
		AppendTags:      setToStringSlice(plan.AppendTags),
		Annotations:     mapToStringMap(plan.Annotations),
		NotifyChannels:  setToStringSlice(plan.NotifyChannels),
		NotifyRecovered: boolToInt(plan.NotifyRecovered.ValueBool()),
		NotifyRuleIDs:   setToInt64Slice(plan.NotifyRuleIDs),
		RunbookURL:      plan.RunbookURL.ValueString(),
	}

	updated, err := r.client.UpdateAlertRule(ctx, plan.BusiGroupID.ValueInt64(), id, apiRule, extra)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert rule, got error: %s", err))
		return
	}

	r.refreshState(ctx, &plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert rule ID: %s", err))
		return
	}

	err = r.client.DeleteAlertRules(ctx, state.BusiGroupID.ValueInt64(), []int64{id})
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert rule, got error: %s", err))
		return
	}
}

func parseImportID(id string) (busiGroupID int64, alertRuleID int64, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected import ID in format busi_group_id:id, got: %s", id)
	}

	busiGroupID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil || busiGroupID <= 0 {
		return 0, 0, fmt.Errorf("invalid busi_group_id: %s", parts[0])
	}

	alertRuleID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || alertRuleID <= 0 {
		return 0, 0, fmt.Errorf("invalid alert rule id: %s", parts[1])
	}

	return busiGroupID, alertRuleID, nil
}

func (r *AlertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	busiGroupID, alertRuleID, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("busi_group_id"), busiGroupID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(alertRuleID, 10))...)
}

func (r *AlertRuleResource) refreshState(ctx context.Context, state *AlertRuleResourceModel, rule *client.AlertRule) {
	state.Name = types.StringValue(rule.Name)
	state.DatasourceType = types.StringValue(rule.DatasourceType)
	state.Disabled = types.BoolValue(rule.Disabled != 0)
	state.Severity = types.Int64Value(rule.Severity)
	state.RunbookURL = types.StringValue(rule.RunbookURL)
	state.CreateAt = types.Int64Value(rule.CreateAt)
	state.CreateBy = types.StringValue(rule.CreateBy)
	state.UpdateAt = types.Int64Value(rule.UpdateAt)
	state.UpdateBy = types.StringValue(rule.UpdateBy)

	if len(rule.DatasourceIDs) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, rule.DatasourceIDs)
		if !diags.HasError() {
			state.DatasourceIDs = setValue
		}
	} else {
		state.DatasourceIDs = types.SetNull(types.Int64Type)
	}

	if len(rule.AppendTags) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.StringType, rule.AppendTags)
		if !diags.HasError() {
			state.AppendTags = setValue
		}
	} else {
		state.AppendTags = types.SetNull(types.StringType)
	}

	if len(rule.Annotations) > 0 {
		mapValue, diags := types.MapValueFrom(ctx, types.StringType, rule.Annotations)
		if !diags.HasError() {
			state.Annotations = mapValue
		}
	} else {
		state.Annotations = types.MapNull(types.StringType)
	}

	if len(rule.NotifyChannels) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.StringType, rule.NotifyChannels)
		if !diags.HasError() {
			state.NotifyChannels = setValue
		}
	} else {
		state.NotifyChannels = types.SetNull(types.StringType)
	}

	if len(rule.NotifyRuleIDs) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, rule.NotifyRuleIDs)
		if !diags.HasError() {
			state.NotifyRuleIDs = setValue
		}
	} else {
		state.NotifyRuleIDs = types.SetNull(types.Int64Type)
	}

	state.NotifyRecovered = types.BoolValue(rule.NotifyRecovered != 0)

	queries, err := client.ParseRuleConfig(rule.RuleConfig)
	if err == nil && len(queries) > 0 {
		queryModels := make([]AlertRuleQueryModel, 0, len(queries))
		for _, q := range queries {
			queryModels = append(queryModels, AlertRuleQueryModel{
				Ref:                types.StringValue(q.Ref),
				PromQL:             types.StringValue(q.PromQL),
				DurationSeconds:    types.Int64Value(q.DurationSeconds),
				ComparisonOperator: types.StringValue(q.ComparisonOperator),
				Threshold:          types.Float64Value(q.Threshold),
			})
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
		if !diags.HasError() {
			state.Queries = listValue
		}
	}
}

func expandQueries(ctx context.Context, list types.List) ([]client.AlertRuleQuery, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var models []AlertRuleQueryModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse queries: %s", diags.Errors()[0].Detail())
	}

	queries := make([]client.AlertRuleQuery, 0, len(models))
	for _, m := range models {
		queries = append(queries, client.AlertRuleQuery{
			Ref:                m.Ref.ValueString(),
			PromQL:             m.PromQL.ValueString(),
			DurationSeconds:    m.DurationSeconds.ValueInt64(),
			ComparisonOperator: m.ComparisonOperator.ValueString(),
			Threshold:          m.Threshold.ValueFloat64(),
		})
	}
	return queries, nil
}

func parseExtraJSON(extraJSON types.String) (map[string]interface{}, error) {
	if extraJSON.IsNull() || extraJSON.IsUnknown() || extraJSON.ValueString() == "" {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(extraJSON.ValueString()), &result); err != nil {
		return nil, fmt.Errorf("extra_json must be a valid JSON object: %s", err)
	}
	return result, nil
}

func setToInt64Slice(set types.Set) []int64 {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var result []int64
	set.ElementsAs(context.Background(), &result, false)
	return result
}

func setToStringSlice(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var result []string
	set.ElementsAs(context.Background(), &result, false)
	return result
}

func mapToStringMap(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	var result map[string]string
	m.ElementsAs(context.Background(), &result, false)
	return result
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "404")
}

type jsonValidator struct{}

func (v jsonValidator) Description(_ context.Context) string {
	return "value must be a valid JSON object"
}

func (v jsonValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid JSON object"
}

func (v jsonValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &m); err != nil {
		resp.Diagnostics.AddError("Invalid extra_json", fmt.Sprintf("extra_json must be a valid JSON object: %s", err))
	}
}
