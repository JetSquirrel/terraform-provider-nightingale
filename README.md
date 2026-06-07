# Terraform Provider for Nightingale

This Terraform provider manages [Nightingale](https://github.com/ccfos/nightingale) alert rules via the Nightingale page-operation API.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

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
