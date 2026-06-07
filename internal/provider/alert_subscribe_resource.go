// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AlertSubscribeResource{}
var _ resource.ResourceWithImportState = &AlertSubscribeResource{}

func NewAlertSubscribeResource() resource.Resource {
	return &AlertSubscribeResource{}
}

type AlertSubscribeResource struct {
	client *client.Client
}

type AlertSubscribeResourceModel struct {
	ID            types.String `tfsdk:"id"`
	BusiGroupID   types.Int64  `tfsdk:"busi_group_id"`
	Name          types.String `tfsdk:"name"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	DatasourceIds types.Set    `tfsdk:"datasource_ids"`
	RuleIds       types.Set    `tfsdk:"rule_ids"`
	Severities    types.Set    `tfsdk:"severities"`
	Tags          types.String `tfsdk:"tags"`
	BusiGroups    types.String `tfsdk:"busi_groups"`
	UserGroupIds  types.Set    `tfsdk:"user_group_ids"`
	NotifyRuleIds types.Set    `tfsdk:"notify_rule_ids"`
	NotifyVersion types.Int64  `tfsdk:"notify_version"`
	CreateAt      types.Int64  `tfsdk:"create_at"`
	CreateBy      types.String `tfsdk:"create_by"`
	UpdateAt      types.Int64  `tfsdk:"update_at"`
	UpdateBy      types.String `tfsdk:"update_by"`
}

func (r *AlertSubscribeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_subscribe"
}

func (r *AlertSubscribeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightingale alert subscription rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Alert subscription rule ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"busi_group_id": schema.Int64Attribute{
				MarkdownDescription: "Nightingale business group ID. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Subscription rule name.",
				Required:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the subscription is disabled. Default is false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"datasource_ids": schema.SetAttribute{
				MarkdownDescription: "Datasource IDs to filter alerts.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"rule_ids": schema.SetAttribute{
				MarkdownDescription: "Alert rule IDs to subscribe to.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"severities": schema.SetAttribute{
				MarkdownDescription: "Severities to match.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"tags": schema.StringAttribute{
				MarkdownDescription: "Tag filter expression.",
				Optional:            true,
			},
			"busi_groups": schema.StringAttribute{
				MarkdownDescription: "Business group filter expression.",
				Optional:            true,
			},
			"user_group_ids": schema.SetAttribute{
				MarkdownDescription: "User group IDs to notify.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"notify_rule_ids": schema.SetAttribute{
				MarkdownDescription: "Notification rule IDs to use.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"notify_version": schema.Int64Attribute{
				MarkdownDescription: "Notify version (1 for new notify rules). Default is 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
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

func (r *AlertSubscribeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *AlertSubscribeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertSubscribeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiSub := r.toAPI(&plan)
	created, err := r.client.CreateAlertSubscribe(ctx, plan.BusiGroupID.ValueInt64(), apiSub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert subscribe: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(created.ID, 10))
	r.refreshState(ctx, &plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertSubscribeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertSubscribeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert subscribe ID: %s", err))
		return
	}

	remote, err := r.client.GetAlertSubscribe(ctx, id)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert subscribe: %s", err))
		return
	}

	r.refreshState(ctx, &state, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertSubscribeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertSubscribeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert subscribe ID: %s", err))
		return
	}

	apiSub := r.toAPI(&plan)
	apiSub.ID = id

	updated, err := r.client.UpdateAlertSubscribe(ctx, plan.BusiGroupID.ValueInt64(), apiSub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert subscribe: %s", err))
		return
	}

	r.refreshState(ctx, &plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertSubscribeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertSubscribeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse alert subscribe ID: %s", err))
		return
	}

	err = r.client.DeleteAlertSubscribes(ctx, state.BusiGroupID.ValueInt64(), []int64{id})
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert subscribe: %s", err))
		return
	}
}

func (r *AlertSubscribeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected busi_group_id:id, got: %s", req.ID))
		return
	}

	busiGroupID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || busiGroupID <= 0 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Invalid busi_group_id: %s", parts[0]))
		return
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Invalid alert subscribe id: %s", parts[1]))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("busi_group_id"), busiGroupID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(id, 10))...)
}

func (r *AlertSubscribeResource) toAPI(state *AlertSubscribeResourceModel) *client.AlertSubscribe {
	return &client.AlertSubscribe{
		GroupId:       state.BusiGroupID.ValueInt64(),
		Name:          state.Name.ValueString(),
		Disabled:      boolToInt(state.Disabled.ValueBool()),
		DatasourceIds: setToInt64Slice(state.DatasourceIds),
		RuleIds:       setToInt64Slice(state.RuleIds),
		Severities:    setToInt64Slice(state.Severities),
		Tags:          state.Tags.ValueString(),
		BusiGroups:    state.BusiGroups.ValueString(),
		UserGroupIds:  setToInt64Slice(state.UserGroupIds),
		NotifyRuleIds: setToInt64Slice(state.NotifyRuleIds),
		NotifyVersion: int(state.NotifyVersion.ValueInt64()),
	}
}

func (r *AlertSubscribeResource) refreshState(ctx context.Context, state *AlertSubscribeResourceModel, sub *client.AlertSubscribe) {
	state.Name = types.StringValue(sub.Name)
	state.Disabled = types.BoolValue(sub.Disabled != 0)
	state.Tags = types.StringValue(sub.Tags)
	state.BusiGroups = types.StringValue(sub.BusiGroups)
	state.NotifyVersion = types.Int64Value(int64(sub.NotifyVersion))
	state.CreateAt = types.Int64Value(sub.CreateAt)
	state.CreateBy = types.StringValue(sub.CreateBy)
	state.UpdateAt = types.Int64Value(sub.UpdateAt)
	state.UpdateBy = types.StringValue(sub.UpdateBy)

	if len(sub.DatasourceIds) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, sub.DatasourceIds)
		if !diags.HasError() {
			state.DatasourceIds = setValue
		}
	} else {
		state.DatasourceIds = types.SetNull(types.Int64Type)
	}

	if len(sub.RuleIds) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, sub.RuleIds)
		if !diags.HasError() {
			state.RuleIds = setValue
		}
	} else {
		state.RuleIds = types.SetNull(types.Int64Type)
	}

	if len(sub.Severities) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, sub.Severities)
		if !diags.HasError() {
			state.Severities = setValue
		}
	} else {
		state.Severities = types.SetNull(types.Int64Type)
	}

	if len(sub.UserGroupIds) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, sub.UserGroupIds)
		if !diags.HasError() {
			state.UserGroupIds = setValue
		}
	} else {
		state.UserGroupIds = types.SetNull(types.Int64Type)
	}

	if len(sub.NotifyRuleIds) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, sub.NotifyRuleIds)
		if !diags.HasError() {
			state.NotifyRuleIds = setValue
		}
	} else {
		state.NotifyRuleIds = types.SetNull(types.Int64Type)
	}
}
