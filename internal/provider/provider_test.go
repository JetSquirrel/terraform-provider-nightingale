// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nightingale": providerserver.NewProtocol6WithError(New("test")()),
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

	p := New("test")().(*NightingaleProvider)
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(configType, values),
	}
}

func TestProviderSchemaTokenSensitive(t *testing.T) {
	p := New("test")().(*NightingaleProvider)
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

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

	p := New("test")().(*NightingaleProvider)
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: providerConfigValue("", "", 0, false),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for missing endpoint and token")
	}
}

func TestProviderConfigurationInvalidEndpoint(t *testing.T) {
	os.Unsetenv("NIGHTINGALE_ENDPOINT")
	os.Unsetenv("NIGHTINGALE_TOKEN")

	p := New("test")().(*NightingaleProvider)
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: providerConfigValue("ftp://invalid", "token", 30, false),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid endpoint scheme")
	}
}

func TestProviderEnvironmentFallback(t *testing.T) {
	os.Setenv("NIGHTINGALE_ENDPOINT", "http://localhost:8080")
	os.Setenv("NIGHTINGALE_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("NIGHTINGALE_ENDPOINT")
		os.Unsetenv("NIGHTINGALE_TOKEN")
	}()

	p := New("test")().(*NightingaleProvider)
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: providerConfigValue("", "", 0, false),
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error when using environment variables: %s", resp.Diagnostics.Errors())
	}
}

func TestProviderResourceRegistration(t *testing.T) {
	p := New("test")().(*NightingaleProvider)
	resources := p.Resources(context.Background())
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
		res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "nightingale"}, &metaResp)
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
