# nightingale_alert_rule

Manages a Nightingale alert rule.

## Example Usage

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

## Schema

### Required

- `busi_group_id` (Number) Nightingale business group ID that owns the alert rule. Changing this forces a new resource.
- `datasource_type` (String) Nightingale datasource type, for example `prometheus`.
- `name` (String) Alert rule name.
- `queries` (Attributes List) Alert query definitions. (see [below for nested schema](#nestedatt--queries))

### Optional

- `annotations` (Map of String) User-facing annotations/metadata.
- `append_tags` (Set of String) Tags appended to generated alert events.
- `disabled` (Boolean) Whether the alert rule is disabled. Default is `false`.
- `datasource_ids` (Set of Number) Datasource IDs used by the rule.
- `extra_json` (String) JSON object merged into API payload for Nightingale-version-specific fields.
- `notify_channels` (Set of String) Notification channels if supported by the target Nightingale version.
- `notify_recovered` (Boolean) Whether to notify on recovery.
- `notify_rule_ids` (Set of Number) Notification rule IDs.
- `runbook_url` (String) Optional runbook URL if supported/mapped through annotations.
- `severity` (Number) Nightingale alert severity.

### Read-Only

- `create_at` (Number) Remote creation timestamp.
- `create_by` (String) Remote creator.
- `id` (String) Nightingale alert rule ID.
- `update_at` (Number) Remote update timestamp.
- `update_by` (String) Remote updater.

<a id="nestedatt--queries"></a>
### Nested Schema for `queries`

Required:

- `promql` (String) PromQL expression.

Optional:

- `comparison_operator` (String) Operator if Nightingale version uses threshold conditions outside PromQL.
- `duration_seconds` (Number) Evaluation duration/for time.
- `ref` (String) Query ref, for example `A`.
- `threshold` (Number) Threshold if Nightingale version uses threshold conditions outside PromQL.

## Import

Import is supported using the following syntax:

```shell
terraform import nightingale_alert_rule.example <busi_group_id>:<alert_rule_id>
```
