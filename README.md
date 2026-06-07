# Terraform Provider for Nightingale

[![Go Report Card](https://goreportcard.com/badge/github.com/JetSquirrel/terraform-provider-nightingale)](https://goreportcard.com/report/github.com/JetSquirrel/terraform-provider-nightingale)

This Terraform provider manages [Nightingale](https://github.com/ccfos/nightingale) resources via the Nightingale page-operation API.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Supported Resources

| Resource | Description | Import Format |
|----------|-------------|---------------|
| [`nightingale_alert_rule`](docs/resources/alert_rule.md) | Alert rules with PromQL queries | `busi_group_id:id` |
| [`nightingale_notify_rule`](docs/resources/notify_rule.md) | Notification rules with channel configs | `id` |
| [`nightingale_alert_subscribe`](docs/resources/alert_subscribe.md) | Alert subscription rules | `busi_group_id:id` |

## Quick Start

### Provider Configuration

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

### Environment Variables

All provider attributes can be set via environment variables:

| Attribute | Environment Variable | Required |
|-----------|---------------------|----------|
| `endpoint` | `NIGHTINGALE_ENDPOINT` | yes |
| `token` | `NIGHTINGALE_TOKEN` | yes |
| `timeout_seconds` | `NIGHTINGALE_TIMEOUT_SECONDS` | no (default: 30) |
| `insecure_skip_tls_verify` | `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` | no (default: false) |

### Resource Examples

#### Alert Rule

```terraform
resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = 1
  name            = "High CPU usage"
  datasource_type = "prometheus"
  datasource_ids  = [1]
  severity        = 2

  queries = [{
    ref              = "A"
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100"
    duration_seconds = 300
  }]

  annotations = {
    summary     = "High CPU usage on {{ $labels.ident }}"
    description = "CPU usage has been high for 5 minutes."
  }

  append_tags = [
    "managed_by=terraform",
  ]
}
```

#### Notification Rule

```terraform
resource "nightingale_notify_rule" "email_ops" {
  name           = "Email OPS"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}
```

#### Alert Subscription

```terraform
resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id   = 1
  name            = "OPS Critical"
  rule_ids        = [10, 11]
  severities      = [1, 2]
  user_group_ids  = [5]
  notify_rule_ids = [3]
}
```

### Import

```shell
terraform import nightingale_alert_rule.high_cpu 1:123
terraform import nightingale_notify_rule.email_ops 456
terraform import nightingale_alert_subscribe.ops_critical 1:789
```

## Building from Source

```shell
git clone https://github.com/JetSquirrel/terraform-provider-nightingale.git
cd terraform-provider-nightingale
go build -v ./...
```

Install the provider locally:

```shell
go install
```

## Testing

Run unit tests:

```shell
go test ./...
```

Run acceptance tests (requires a live Nightingale instance):

```shell
export NIGHTINGALE_ACC=1
export NIGHTINGALE_ENDPOINT="https://n9e.example.com"
export NIGHTINGALE_TOKEN="your-token"
export NIGHTINGALE_BUSI_GROUP_ID=1
go test ./... -run 'TestAcc'
```

## Documentation Generation

Documentation is generated from schema definitions:

```shell
make generate
```

## Documentation

- [Provider Configuration](docs/index.md)
- [Alert Rule](docs/resources/alert_rule.md)
- [Notification Rule](docs/resources/notify_rule.md)
- [Alert Subscription](docs/resources/alert_subscribe.md)

## License

MPL-2.0
