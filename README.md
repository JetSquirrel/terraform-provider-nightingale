# Terraform Provider for Nightingale

This Terraform provider manages [Nightingale](https://github.com/ccfos/nightingale) resources via the Nightingale page-operation API.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Resources

- `nightingale_alert_rule` — Alert rules
- `nightingale_notify_rule` — Notification rules
- `nightingale_alert_subscribe` — Alert subscription rules

## Building the Provider

```shell
go build -v ./...
```

## Using the Provider

```terraform
provider "nightingale" {
  endpoint = "https://n9e.example.com"
  token    = var.nightingale_token
}

resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = 1
  name            = "High CPU usage"
  datasource_type = "prometheus"
  severity        = 2

  queries = [{
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100"
    duration_seconds = 300
  }]

  annotations = {
    summary = "High CPU usage on {{ $labels.ident }}"
  }
}

resource "nightingale_notify_rule" "email_ops" {
  name           = "Email OPS"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}

resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = 1
  name          = "OPS Critical"
  rule_ids      = [10, 11]
  severities    = [1, 2]
  user_group_ids = [5]
  notify_rule_ids = [3]
}
```

## Developing the Provider

To compile the provider:

```shell
go install
```

To run the test suite:

```shell
go test ./...
```

To generate or update documentation:

```shell
make generate
```

## Documentation

Full documentation is available in the `docs/` directory and follows the Terraform provider documentation conventions.
