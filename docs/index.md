# Nightingale Provider

The Nightingale provider allows you to manage [Nightingale](https://github.com/ccfos/nightingale) alert rules using Terraform.

## Example Usage

```terraform
provider "nightingale" {
  endpoint = var.nightingale_endpoint
  token    = var.nightingale_token
}
```

## Schema

### Required

- `endpoint` (String) Base URL for the Nightingale center API. May be set via `NIGHTINGALE_ENDPOINT` environment variable.
- `token` (String, Sensitive) User token sent as `X-User-Token`. May be set via `NIGHTINGALE_TOKEN` environment variable.

### Optional

- `insecure_skip_tls_verify` (Boolean) Skip TLS certificate verification. Default is `false`. May be set via `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` environment variable.
- `timeout_seconds` (Number) HTTP timeout in seconds. Default is `30`. May be set via `NIGHTINGALE_TIMEOUT_SECONDS` environment variable.
