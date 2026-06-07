// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/JetSquirrel/terraform-provider-nightingale/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure NightingaleProvider satisfies various provider interfaces.
var _ provider.Provider = &NightingaleProvider{}

// NightingaleProvider defines the provider implementation.
type NightingaleProvider struct {
	version string
}

// NightingaleProviderModel describes the provider data model.
type NightingaleProviderModel struct {
	Endpoint              types.String `tfsdk:"endpoint"`
	Token                 types.String `tfsdk:"token"`
	TimeoutSeconds        types.Int64  `tfsdk:"timeout_seconds"`
	InsecureSkipTLSVerify types.Bool   `tfsdk:"insecure_skip_tls_verify"`
}

func (p *NightingaleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nightingale"
	resp.Version = p.version
}

func (p *NightingaleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base URL for the Nightingale center API. May be set via NIGHTINGALE_ENDPOINT environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "User token sent as X-User-Token. May be set via NIGHTINGALE_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "HTTP timeout in seconds. Default is 30. May be set via NIGHTINGALE_TIMEOUT_SECONDS environment variable.",
				Optional:            true,
			},
			"insecure_skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Default is false. May be set via NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *NightingaleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NightingaleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := getConfigValue(data.Endpoint, "NIGHTINGALE_ENDPOINT")
	token := getConfigValue(data.Token, "NIGHTINGALE_TOKEN")
	timeoutSeconds := int64(30)
	insecureSkipTLSVerify := false

	if !data.TimeoutSeconds.IsNull() && !data.TimeoutSeconds.IsUnknown() {
		timeoutSeconds = data.TimeoutSeconds.ValueInt64()
	} else if v := os.Getenv("NIGHTINGALE_TIMEOUT_SECONDS"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			timeoutSeconds = parsed
		}
	}

	if !data.InsecureSkipTLSVerify.IsNull() && !data.InsecureSkipTLSVerify.IsUnknown() {
		insecureSkipTLSVerify = data.InsecureSkipTLSVerify.ValueBool()
	} else if v := os.Getenv("NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			insecureSkipTLSVerify = parsed
		}
	}

	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"The endpoint attribute or NIGHTINGALE_ENDPOINT environment variable must be set.",
		)
		return
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		resp.Diagnostics.AddError(
			"Invalid Endpoint Configuration",
			fmt.Sprintf("The endpoint must have http or https scheme: %s", endpoint),
		)
		return
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Token Configuration",
			"The token attribute or NIGHTINGALE_TOKEN environment variable must be set.",
		)
		return
	}

	userAgent := fmt.Sprintf("terraform-provider-nightingale/%s", p.version)
	c, err := client.New(endpoint, token, int(timeoutSeconds), insecureSkipTLSVerify, userAgent)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Nightingale Client",
			fmt.Sprintf("Error creating client: %s", err),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *NightingaleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAlertRuleResource,
	}
}

func (p *NightingaleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NightingaleProvider{
			version: version,
		}
	}
}

func getConfigValue(attr types.String, envVar string) string {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueString()
	}
	return os.Getenv(envVar)
}
