// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func newTestProvider(t *testing.T) *NightingaleProvider {
	t.Helper()
	p, ok := New("test")().(*NightingaleProvider)
	if !ok {
		t.Fatal("expected *NightingaleProvider")
	}
	return p
}

func providerConfigValue(endpoint, token string, timeout int64, insecure bool) tfsdk.Config {
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"endpoint":                 tftypes.String,
			"token":                    tftypes.String,
			"timeout_seconds":          tftypes.Number,
			"insecure_skip_tls_verify": tftypes.Bool,
		},
		OptionalAttributes: map[string]struct{}{
			"endpoint":                 {},
			"token":                    {},
			"timeout_seconds":          {},
			"insecure_skip_tls_verify": {},
		},
	}

	values := map[string]tftypes.Value{}
	if endpoint != "" {
		values["endpoint"] = tftypes.NewValue(tftypes.String, endpoint)
	} else {
		values["endpoint"] = tftypes.NewValue(tftypes.String, nil)
	}
	if token != "" {
		values["token"] = tftypes.NewValue(tftypes.String, token)
	} else {
		values["token"] = tftypes.NewValue(tftypes.String, nil)
	}
	if timeout > 0 {
		values["timeout_seconds"] = tftypes.NewValue(tftypes.Number, float64(timeout))
	} else {
		values["timeout_seconds"] = tftypes.NewValue(tftypes.Number, nil)
	}
	values["insecure_skip_tls_verify"] = tftypes.NewValue(tftypes.Bool, insecure)

	p := New("test")()
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(configType, values),
	}
}

func TestProviderSchemaTokenSensitive(t *testing.T) {
	p := newTestProvider(t)
	var resp provider.SchemaResponse
	p.Schema(t.Context(), provider.SchemaRequest{}, &resp)

	if resp.Schema.Attributes["token"] == nil {
		t.Fatal("token attribute not found in schema")
	}
	if !resp.Schema.Attributes["token"].IsSensitive() {
		t.Error("expected token to be marked sensitive")
	}
}

func TestProviderConfigurationMissingValues(t *testing.T) {
	os.Unsetenv("NIGHTINGALE_ENDPOINT")
	os.Unsetenv("NIGHTINGALE_TOKEN")

	p := newTestProvider(t)
	var resp provider.ConfigureResponse
	p.Configure(t.Context(), provider.ConfigureRequest{
		Config: providerConfigValue("", "", 0, false),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for missing endpoint and token")
	}
}

func TestProviderConfigurationInvalidEndpoint(t *testing.T) {
	os.Unsetenv("NIGHTINGALE_ENDPOINT")
	os.Unsetenv("NIGHTINGALE_TOKEN")

	p := newTestProvider(t)
	var resp provider.ConfigureResponse
	p.Configure(t.Context(), provider.ConfigureRequest{
		Config: providerConfigValue("ftp://invalid", "token", 30, false),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid endpoint scheme")
	}
}

func TestProviderEnvironmentFallback(t *testing.T) {
	t.Setenv("NIGHTINGALE_ENDPOINT", "http://localhost:8080")
	t.Setenv("NIGHTINGALE_TOKEN", "env-token")

	p := newTestProvider(t)
	var resp provider.ConfigureResponse
	p.Configure(t.Context(), provider.ConfigureRequest{
		Config: providerConfigValue("", "", 0, false),
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error when using environment variables: %s", resp.Diagnostics.Errors())
	}
}

func TestProviderResourceRegistration(t *testing.T) {
	p := newTestProvider(t)
	resources := p.Resources(t.Context())
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	expectedTypes := map[string]bool{
		"nightingale_alert_rule":      false,
		"nightingale_notify_rule":     false,
		"nightingale_alert_subscribe": false,
	}

	for _, factory := range resources {
		res := factory()
		var metaResp resource.MetadataResponse
		res.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "nightingale"}, &metaResp)
		if _, ok := expectedTypes[metaResp.TypeName]; !ok {
			t.Errorf("unexpected resource type: %s", metaResp.TypeName)
		}
		expectedTypes[metaResp.TypeName] = true
	}

	for name, found := range expectedTypes {
		if !found {
			t.Errorf("missing resource type: %s", name)
		}
	}
}
