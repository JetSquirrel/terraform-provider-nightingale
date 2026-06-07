# Terraform Provider for Nightingale

[![Go Report Card](https://goreportcard.com/badge/github.com/JetSquirrel/terraform-provider-nightingale)](https://goreportcard.com/report/github.com/JetSquirrel/terraform-provider-nightingale)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

[English](README.md) | [中文](README_CN.md)

Terraform provider for managing [Nightingale](https://github.com/ccfos/nightingale) (n9e) monitoring resources.

## Features

- Manage alert rules with PromQL queries
- Configure notification rules with multiple channels
- Set up alert subscriptions for teams
- Full CRUD support with import capability
- Compatible with Nightingale v9.x API

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Nightingale v8.0+ (tested with v9.x)

## Installation

### From Terraform Registry

```terraform
terraform {
  required_providers {
    nightingale = {
      source  = "JetSquirrel/nightingale"
      version = "0.0.1"
    }
  }
}
```

### From Source

```shell
git clone https://github.com/JetSquirrel/terraform-provider-nightingale.git
cd terraform-provider-nightingale
go install
```

## Quick Start

### 1. Configure Provider

```terraform
provider "nightingale" {
  endpoint = "https://n9e.example.com"
  token    = var.nightingale_token
}
```

Or use environment variables:

```shell
export NIGHTINGALE_ENDPOINT="https://n9e.example.com"
export NIGHTINGALE_TOKEN="your-api-token"
```

### 2. Get API Token

1. Log in to Nightingale web UI
2. Go to **Profile** (top-right avatar) > **Token Management**
3. Click **Create Token**
4. Copy the generated token

> **Note:** Ensure `[HTTP.TokenAuth] Enable = true` is set in Nightingale's `config.toml`.

### 3. Create Resources

```terraform
# Alert rule for high CPU usage
resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = 1
  name            = "High CPU Usage"
  datasource_type = "prometheus"
  datasource_ids  = [1]
  severity        = 2

  queries = [{
    ref              = "A"
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100 > 80"
    duration_seconds = 300
  }]

  annotations = {
    summary     = "High CPU on {{ $labels.ident }}"
    description = "CPU usage exceeded 80% for 5 minutes."
  }

  append_tags = ["managed_by=terraform"]
}
```

## Supported Resources

| Resource | Description | Import Format |
|----------|-------------|---------------|
| [`nightingale_alert_rule`](docs/resources/alert_rule.md) | Alert rules with PromQL queries | `busi_group_id:id` |
| [`nightingale_notify_rule`](docs/resources/notify_rule.md) | Notification rules | `id` |
| [`nightingale_alert_subscribe`](docs/resources/alert_subscribe.md) | Alert subscriptions | `busi_group_id:id` |

## Provider Configuration

| Attribute | Environment Variable | Required | Default | Description |
|-----------|---------------------|----------|---------|-------------|
| `endpoint` | `NIGHTINGALE_ENDPOINT` | Yes | - | Nightingale API URL |
| `token` | `NIGHTINGALE_TOKEN` | Yes | - | API token (X-User-Token) |
| `timeout_seconds` | `NIGHTINGALE_TIMEOUT_SECONDS` | No | 30 | HTTP timeout |
| `insecure_skip_tls_verify` | `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` | No | false | Skip TLS verification |

## Examples

See the [examples](examples/) directory for complete configurations:

- [Provider setup](examples/provider/)
- [Complete example](examples/complete/) - Multiple resources working together
- [Individual resources](examples/resources/)

## Import Existing Resources

```shell
# Alert rule
terraform import nightingale_alert_rule.example 1:123

# Notification rule
terraform import nightingale_notify_rule.example 456

# Alert subscription
terraform import nightingale_alert_subscribe.example 1:789
```

## Development

### Build

```shell
make build
```

### Test

```shell
# Unit tests
make test

# Acceptance tests (requires live Nightingale instance)
export TF_ACC=1
export NIGHTINGALE_ENDPOINT="http://localhost:17000"
export NIGHTINGALE_TOKEN="your-token"
make testacc
```

### Generate Documentation

```shell
make generate
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MPL-2.0](LICENSE)

## Links

- [Nightingale Project](https://github.com/ccfos/nightingale)
- [Nightingale Documentation](https://flashcat.cloud/docs/)
- [Terraform Registry](https://registry.terraform.io/providers/JetSquirrel/nightingale)
