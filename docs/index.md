---
page_title: "Nightingale Provider"
description: |-
  Terraform provider for managing Nightingale monitoring resources.
---

# Nightingale Provider

The Nightingale provider enables Terraform to manage [Nightingale](https://github.com/ccfos/nightingale) (n9e) monitoring resources including alert rules, notification rules, and alert subscriptions.

## Example Usage

```terraform
terraform {
  required_providers {
    nightingale = {
      source  = "JetSquirrel/nightingale"
      version = "~> 0.1"
    }
  }
}

provider "nightingale" {
  endpoint = "https://n9e.example.com"
  token    = var.nightingale_token
}
```

## Authentication

The provider requires an API token for authentication. To obtain a token:

1. Log in to Nightingale web UI
2. Navigate to **Profile** > **Token Management**
3. Create a new token

Ensure `[HTTP.TokenAuth] Enable = true` is configured in Nightingale's `config.toml`.

## Schema

### Required

- `endpoint` (String) Nightingale API endpoint URL. Can be set via `NIGHTINGALE_ENDPOINT` environment variable.
- `token` (String, Sensitive) API token for authentication. Can be set via `NIGHTINGALE_TOKEN` environment variable.

### Optional

- `timeout_seconds` (Number) HTTP request timeout in seconds. Default: `30`. Can be set via `NIGHTINGALE_TIMEOUT_SECONDS` environment variable.
- `insecure_skip_tls_verify` (Boolean) Skip TLS certificate verification. Default: `false`. Can be set via `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` environment variable.

## Resources

- [nightingale_alert_rule](resources/alert_rule.md) - Manage alert rules
- [nightingale_notify_rule](resources/notify_rule.md) - Manage notification rules
- [nightingale_alert_subscribe](resources/alert_subscribe.md) - Manage alert subscriptions
