# Terraform Provider for Nightingale — Alert as Code Spec

## Purpose

Build a Terraform provider named `nightingale` for [CCF Nightingale](https://github.com/ccfos/nightingale), starting with an alert-as-code implementation. The first production resource must manage Nightingale alert rules through the Nightingale page-operation API documented at <https://n9e.github.io/docs/usecase/api/>.

This spec is written for an implementation agent. Treat it as the acceptance contract for the first provider milestone.

## Source references

- Nightingale repository: <https://github.com/ccfos/nightingale>
- Nightingale API guide: <https://n9e.github.io/docs/usecase/api/>
- Nightingale API authentication requirements:
  - Enable `HTTP.TokenAuth` in Nightingale config.
  - Send `X-User-Token` on page-operation requests.
  - Send `Content-Type: application/json` for JSON requests.
  - A successful API call requires both HTTP status `200` and a response JSON object whose `err` field is empty.
- Nightingale API discovery guidance:
  - The official docs state that page-operation APIs mirror browser UI operations and recommend using Chrome/DevTools Network requests to verify exact paths and payloads for the deployed Nightingale version.
  - The agent must verify alert-rule paths and payload shape from the target Nightingale source/UI before finalizing CRUD.
- Alert rule route family to verify in Nightingale source/UI:
  - `/api/n9e/busi-group/:id/alert-rules`

## Current repository state

The repository is currently a Terraform Plugin Framework scaffold. The implementation agent must replace scaffold identity, examples, and behavior with a real Nightingale provider.

Known scaffold artifacts to replace or remove:

- Provider type name `scaffolding`.
- Registry address `registry.terraform.io/hashicorp/scaffolding`.
- Example provider/resource/data source/action/function/ephemeral implementations.
- Example docs under `docs/` and `examples/`.
- Test factory provider name `scaffolding`.
- Module path `github.com/hashicorp/terraform-provider-scaffolding-framework` unless project ownership intentionally keeps another path.

## Non-goals for milestone 1

Do not implement every Nightingale resource in the first milestone. The first milestone is focused on provider infrastructure and `nightingale_alert_rule`.

Out of scope for milestone 1 unless needed by tests:

- Dashboards.
- Business group CRUD.
- Datasource CRUD.
- Notification rule CRUD.
- Mute/subscription/event-pipeline CRUD.
- Acceptance tests against a live Nightingale instance.

## Provider requirements

### Provider identity

- Terraform provider type name: `nightingale`.
- Default local binary name should follow Terraform conventions for the chosen module/registry namespace.
- Replace all user-facing scaffold names with Nightingale names.
- Use Terraform Plugin Framework idioms already present in the repository.

### Provider configuration schema

Implement these provider attributes:

| Attribute | Type | Required | Sensitive | Environment fallback | Description |
| --- | --- | --- | --- | --- | --- |
| `endpoint` | string | yes unless env set | no | `NIGHTINGALE_ENDPOINT` | Base URL for the Nightingale center API, for example `https://n9e.example.com`. |
| `token` | string | yes unless env set | yes | `NIGHTINGALE_TOKEN` | User token sent as `X-User-Token`. |
| `timeout_seconds` | number | no | no | `NIGHTINGALE_TIMEOUT_SECONDS` | HTTP timeout. Default: `30`. |
| `insecure_skip_tls_verify` | bool | no | no | `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` | Optional TLS escape hatch for development only. Default: `false`. |

Validation and normalization:

- Trim trailing slashes from `endpoint`.
- Reject an empty endpoint after config/env fallback.
- Reject an endpoint without `http` or `https` scheme.
- Reject an empty token after config/env fallback.
- Mark `token` as sensitive in schema and diagnostics.
- Do not log token values.

### Provider client

Create an internal client package, recommended path: `internal/client`.

Client requirements:

- Holds base endpoint, token, configured `http.Client`, and optional user agent.
- Adds these headers to page-operation requests:
  - `X-User-Token: <token>`
  - `Content-Type: application/json` for requests with JSON bodies
  - `Accept: application/json`
  - Provider user agent, for example `terraform-provider-nightingale/<version>`
- Encodes JSON request bodies.
- Decodes JSON response envelopes.
- Treats non-`200` HTTP status codes as errors and includes method, path, status code, and a bounded response body preview.
- Treats HTTP `200` with non-empty JSON `err` as an API error.
- Handles `dat` as a generic JSON field that can be decoded into typed structs.
- Provides typed methods for alert rules instead of scattering raw endpoint strings in resource logic.

Recommended response envelope:

```go
type Envelope struct {
    Dat json.RawMessage `json:"dat"`
    Err string          `json:"err"`
}
```

## `nightingale_alert_rule` resource requirements

### Terraform name

Resource type: `nightingale_alert_rule`.

### Lifecycle behavior

Implement full Terraform lifecycle:

- Create: create a Nightingale alert rule and store the returned ID.
- Read: fetch remote state, update Terraform state, and remove state if the rule no longer exists.
- Update: update the Nightingale alert rule in place.
- Delete: delete the Nightingale alert rule and tolerate already-deleted remote objects.
- Import: support importing by `busi_group_id:id`.

### API route assumptions to verify

The expected route family is:

- Collection: `/api/n9e/busi-group/{busi_group_id}/alert-rules`
- Item operations may use collection body IDs or item paths depending on Nightingale version.

The agent must verify exact methods and payloads from the Nightingale source or browser DevTools before implementing final CRUD. If the deployed/source Nightingale version differs from this route family, update the client and documentation with the verified route details.

### Schema

The schema should cover common Nightingale alert rule fields while keeping the first milestone maintainable. Use strongly typed attributes for common fields and an escape hatch for version-specific fields.

Required attributes:

| Attribute | Type | Description |
| --- | --- | --- |
| `busi_group_id` | number/int64 | Nightingale business group ID that owns the alert rule. Force replacement if changed. |
| `name` | string | Alert rule name. |
| `datasource_type` | string | Nightingale datasource type, for example `prometheus`. |
| `queries` | list(object) | Alert query definitions. At minimum support Prometheus/PromQL rules. |

Computed attributes:

| Attribute | Type | Description |
| --- | --- | --- |
| `id` | string | Nightingale alert rule ID. |
| `create_at` | number/int64 | Remote creation timestamp when returned by Nightingale. |
| `create_by` | string | Remote creator when returned by Nightingale. |
| `update_at` | number/int64 | Remote update timestamp when returned by Nightingale. |
| `update_by` | string | Remote updater when returned by Nightingale. |

Optional or optional+computed attributes:

| Attribute | Type | Default | Description |
| --- | --- | --- | --- |
| `disabled` | bool | `false` | Whether the alert rule is disabled. |
| `severity` | number/int64 | implementation default or required if Nightingale requires it | Nightingale alert severity. |
| `datasource_ids` | set(number/int64) | none | Datasource IDs used by the rule. |
| `append_tags` | set(string) | none | Tags appended to generated alert events. |
| `annotations` | map(string) | none | User-facing annotations/metadata. |
| `notify_rule_ids` | set(number/int64) | none | Notification rule IDs. |
| `notify_recovered` | bool | Nightingale default | Whether to notify on recovery. |
| `notify_channels` | set(string) | none | Notification channels if supported by the target Nightingale version. |
| `runbook_url` | string | none | Optional runbook URL if supported/mapped through annotations. |
| `extra_json` | string | none | JSON object merged into API payload for Nightingale-version-specific fields. |

Recommended nested `queries` object for first milestone:

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `ref` | string | optional | Query ref, for example `A`. |
| `promql` | string | yes for Prometheus | PromQL expression. |
| `duration_seconds` | number/int64 | optional | Evaluation duration/for time. |
| `comparison_operator` | string | optional | Operator if Nightingale version uses threshold conditions outside PromQL. |
| `threshold` | number | optional | Threshold if Nightingale version uses threshold conditions outside PromQL. |

Implementation notes:

- Use Terraform `Set` types where ordering is irrelevant.
- Preserve server defaults to avoid perpetual diffs.
- Validate JSON in `extra_json` at plan time.
- Validate IDs are positive where applicable.
- Use plan modifiers for computed `id` and server-managed timestamps.
- Prefer explicit typed fields over `extra_json`; `extra_json` is only an escape hatch.

### Import format

Import ID format:

```text
<busi_group_id>:<alert_rule_id>
```

Import behavior:

- Parse both segments as positive integers.
- Set `busi_group_id` and `id` in state.
- Trigger read to populate the remaining attributes.
- Return clear diagnostics for malformed import IDs.

### Drift handling

Read behavior must reconcile remote state:

- If remote alert rule exists, refresh all mapped Terraform attributes.
- If remote alert rule is missing, call `resp.State.RemoveResource(ctx)`.
- If remote response has unknown or unsupported fields, preserve managed fields and ignore unknown fields unless represented by `extra_json`.
- Normalize semantically equivalent values, especially empty lists/maps and server defaults.

## Tests

### Unit tests

Add unit tests with `httptest.Server` for the client and resource mapping.

Required client test cases:

- Sends `X-User-Token` and JSON headers.
- Builds paths relative to configured endpoint without duplicate slashes.
- Returns success when HTTP status is `200` and `err` is empty.
- Returns error on non-`200` status with response preview.
- Returns error on HTTP `200` with non-empty `err`.
- Handles malformed JSON response.
- Does not leak token in error strings.

Required resource test cases:

- Create maps Terraform plan to Nightingale request payload and persists returned ID.
- Read updates state from remote payload.
- Read removes state when remote rule is absent.
- Update sends changed fields.
- Delete sends delete request and tolerates not found.
- Import parses `busi_group_id:id`.
- Import rejects malformed IDs.
- `extra_json` rejects invalid JSON.

### Provider/schema tests

- Provider can instantiate under Terraform Plugin Framework test server as `nightingale`.
- Provider schema marks `token` sensitive.
- Provider configuration supports environment fallback.
- Resource schema contains `nightingale_alert_rule`.

### Acceptance tests

Acceptance tests are optional in milestone 1. If added, gate them behind environment variables:

- `NIGHTINGALE_ACC=1`
- `NIGHTINGALE_ENDPOINT`
- `NIGHTINGALE_TOKEN`
- `NIGHTINGALE_BUSI_GROUP_ID`
- Optional `NIGHTINGALE_DATASOURCE_ID`

Do not run live acceptance tests by default in CI.

## Documentation and examples

Update docs and examples for the Nightingale provider.

Required files:

- `README.md`
- `docs/index.md`
- `docs/resources/alert_rule.md`
- `examples/provider/provider.tf`
- `examples/resources/nightingale_alert_rule/resource.tf`
- `examples/resources/nightingale_alert_rule/import.sh`

Minimum provider example:

```hcl
provider "nightingale" {
  endpoint = var.nightingale_endpoint
  token    = var.nightingale_token
}
```

Minimum alert rule example:

```hcl
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

## Code quality expectations

- Keep API concerns in `internal/client`; keep Terraform schema/state concerns in `internal/provider`.
- Avoid try/catch-style wrappers around imports.
- Do not log secrets.
- Prefer typed request/response structs.
- Keep tests deterministic and local by default.
- Run `gofmt` on changed Go files.
- Run `go test ./...` before committing.
- Regenerate docs if the repository has a docs generation target available.

## Milestone 1 acceptance checklist

The milestone is complete when all of the following are true:

- [ ] Provider serves as `nightingale`.
- [ ] Provider accepts endpoint/token config and environment fallbacks.
- [ ] Provider client authenticates with `X-User-Token`.
- [ ] Provider client checks both HTTP status and Nightingale JSON `err`.
- [ ] `nightingale_alert_rule` is registered.
- [ ] `nightingale_alert_rule` implements create/read/update/delete/import.
- [ ] Remote deletion removes Terraform state.
- [ ] Client and resource unit tests cover success and error paths.
- [ ] Scaffold examples and docs are replaced with Nightingale docs.
- [ ] `go test ./...` passes locally.
- [ ] Changes are committed on the current branch.
