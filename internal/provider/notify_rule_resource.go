// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NotifyRuleResource{}
var _ resource.ResourceWithImportState = &NotifyRuleResource{}

func NewNotifyRuleResource() resource.Resource {
	return &NotifyRuleResource{}
}

type NotifyRuleResource struct {
	client *client.Client
}

type NotifyConfigModel struct {
	ChannelID  types.Int64 `tfsdk:"channel_id"`
	TemplateID types.Int64 `tfsdk:"template_id"`
	Params     types.Map   `tfsdk:"params"`
}

type NotifyRuleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Enable        types.Bool   `tfsdk:"enable"`
	UserGroupIds  types.Set    `tfsdk:"user_group_ids"`
	NotifyConfigs types.List   `tfsdk:"notify_configs"`
	CreateAt      types.Int64  `tfsdk:"create_at"`
	CreateBy      types.String `tfsdk:"create_by"`
	UpdateAt      types.Int64  `tfsdk:"update_at"`
	UpdateBy      types.String `tfsdk:"update_by"`
}

func (r *NotifyRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notify_rule"
}

func (r *NotifyRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightingale notification rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Notification rule ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Notification rule name.",
				Required:            true,
			},
			"enable": schema.BoolAttribute{
				MarkdownDescription: "Whether the notification rule is enabled. Default is true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"user_group_ids": schema.SetAttribute{
				MarkdownDescription: "User group IDs associated with this rule.",
				Required:            true,
				ElementType:         types.Int64Type,
			},
			"notify_configs": schema.ListNestedAttribute{
				MarkdownDescription: "Notification channel configurations.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel_id": schema.Int64Attribute{
							MarkdownDescription: "Notification channel ID.",
							Required:            true,
						},
						"template_id": schema.Int64Attribute{
							MarkdownDescription: "Message template ID.",
							Optional:            true,
						},
						"params": schema.MapAttribute{
							MarkdownDescription: "Custom parameters for the notification channel.",
							Optional:            true,
							ElementType:         types.StringType,
						},
					},
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

func (r *NotifyRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotifyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotifyRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configs, err := expandNotifyConfigs(ctx, plan.NotifyConfigs)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	apiRule := &client.NotifyRule{
		Name:          plan.Name.ValueString(),
		Enable:        boolToInt(plan.Enable.ValueBool()),
		UserGroupIds:  setToInt64Slice(plan.UserGroupIds),
		NotifyConfigs: configs,
	}

	created, err := r.client.CreateNotifyRule(ctx, apiRule)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notify rule: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(created.ID, 10))
	r.refreshState(ctx, &plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotifyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotifyRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse notify rule ID: %s", err))
		return
	}

	remote, err := r.client.GetNotifyRule(ctx, id)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notify rule: %s", err))
		return
	}

	r.refreshState(ctx, &state, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotifyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotifyRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse notify rule ID: %s", err))
		return
	}

	configs, err := expandNotifyConfigs(ctx, plan.NotifyConfigs)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	apiRule := &client.NotifyRule{
		ID:            id,
		Name:          plan.Name.ValueString(),
		Enable:        boolToInt(plan.Enable.ValueBool()),
		UserGroupIds:  setToInt64Slice(plan.UserGroupIds),
		NotifyConfigs: configs,
	}

	updated, err := r.client.UpdateNotifyRule(ctx, id, apiRule)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notify rule: %s", err))
		return
	}

	r.refreshState(ctx, &plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotifyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotifyRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse notify rule ID: %s", err))
		return
	}

	err = r.client.DeleteNotifyRules(ctx, []int64{id})
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notify rule: %s", err))
		return
	}
}

func (r *NotifyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected numeric notify rule ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(id, 10))...)
}

func (r *NotifyRuleResource) refreshState(ctx context.Context, state *NotifyRuleResourceModel, rule *client.NotifyRule) {
	state.Name = types.StringValue(rule.Name)
	state.Enable = types.BoolValue(rule.Enable != 0)
	state.CreateAt = types.Int64Value(rule.CreateAt)
	state.CreateBy = types.StringValue(rule.CreateBy)
	state.UpdateAt = types.Int64Value(rule.UpdateAt)
	state.UpdateBy = types.StringValue(rule.UpdateBy)

	if len(rule.UserGroupIds) > 0 {
		setValue, diags := types.SetValueFrom(ctx, types.Int64Type, rule.UserGroupIds)
		if !diags.HasError() {
			state.UserGroupIds = setValue
		}
	} else {
		state.UserGroupIds = types.SetNull(types.Int64Type)
	}

	if len(rule.NotifyConfigs) > 0 {
		configModels := make([]NotifyConfigModel, 0, len(rule.NotifyConfigs))
		for _, cfg := range rule.NotifyConfigs {
			params, _ := types.MapValueFrom(ctx, types.StringType, stringifyMap(cfg.Params))
			configModels = append(configModels, NotifyConfigModel{
				ChannelID:  types.Int64Value(cfg.ChannelID),
				TemplateID: types.Int64Value(cfg.TemplateID),
				Params:     params,
			})
		}
		listValue, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"channel_id":  types.Int64Type,
				"template_id": types.Int64Type,
				"params":      types.MapType{ElemType: types.StringType},
			},
		}, configModels)
		if !diags.HasError() {
			state.NotifyConfigs = listValue
		}
	} else {
		state.NotifyConfigs = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"channel_id":  types.Int64Type,
				"template_id": types.Int64Type,
				"params":      types.MapType{ElemType: types.StringType},
			},
		})
	}
}

func expandNotifyConfigs(ctx context.Context, list types.List) ([]client.NotifyConfig, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var models []NotifyConfigModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse notify_configs: %s", diags.Errors()[0].Detail())
	}

	configs := make([]client.NotifyConfig, 0, len(models))
	for _, m := range models {
		params := mapToStringMap(m.Params)
		// Convert string map to interface{} map for JSON serialization
		ifaceParams := make(map[string]interface{}, len(params))
		for k, v := range params {
			ifaceParams[k] = v
		}
		configs = append(configs, client.NotifyConfig{
			ChannelID:  m.ChannelID.ValueInt64(),
			TemplateID: m.TemplateID.ValueInt64(),
			Params:     ifaceParams,
		})
	}
	return configs, nil
}

func stringifyMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
