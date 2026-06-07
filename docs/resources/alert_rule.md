---
page_title: "nightingale_alert_rule Resource - Nightingale"
description: |-
  Manages a Nightingale alert rule with PromQL queries.
---

# nightingale_alert_rule

Manages a Nightingale alert rule. Alert rules define conditions using PromQL queries that trigger alerts when met.

## Example Usage

### Basic Alert Rule

```terraform
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

### Alert Rule with Notification

```terraform
resource "nightingale_alert_rule" "disk_full" {
  busi_group_id    = 1
  name             = "Disk Nearly Full"
  datasource_type  = "prometheus"
  datasource_ids   = [1]
  severity         = 1
  notify_recovered = true

  queries = [{
    ref              = "A"
    promql           = "node_filesystem_avail_bytes / node_filesystem_size_bytes * 100 < 10"
    duration_seconds = 600
  }]

  annotations = {
    summary     = "Disk nearly full on {{ $labels.ident }}"
    description = "Available disk space is below 10%."
  }

  append_tags     = ["severity=critical", "managed_by=terraform"]
  notify_rule_ids = [nightingale_notify_rule.ops.id]
}
```

## Schema

### Required

- `busi_group_id` (Number) Business group ID that owns the alert rule. Changing this forces a new resource.
- `datasource_type` (String) Datasource type (e.g., `prometheus`).
- `name` (String) Alert rule name.
- `queries` (Block List, Min: 1) Alert query definitions. See [Nested Schema](#nestedblock--queries) below.

### Optional

- `annotations` (Map of String) Key-value annotations for alert context. Supports template variables like `{{ $labels.ident }}`.
- `append_tags` (Set of String) Tags appended to generated alerts.
- `datasource_ids` (Set of Number) Datasource IDs to query.
- `disabled` (Boolean) Whether the rule is disabled. Default: `false`.
- `extra_json` (String) JSON object merged into API payload for version-specific fields.
- `notify_channels` (Set of String) Notification channels (if supported).
- `notify_recovered` (Boolean) Send notification when alert recovers.
- `notify_rule_ids` (Set of Number) Notification rule IDs to trigger.
- `runbook_url` (String) Runbook URL for the alert.
- `severity` (Number) Alert severity level (1=Critical, 2=Warning, 3=Info).

### Read-Only

- `create_at` (Number) Creation timestamp.
- `create_by` (String) Creator username.
- `id` (String) Alert rule ID.
- `update_at` (Number) Last update timestamp.
- `update_by` (String) Last updater username.

<a id="nestedblock--queries"></a>
### Nested Schema for `queries`

Required:

- `promql` (String) PromQL expression.

Optional:

- `comparison_operator` (String) Comparison operator for threshold-based alerting.
- `duration_seconds` (Number) Duration in seconds the condition must be true.
- `ref` (String) Query reference identifier (e.g., `A`, `B`).
- `threshold` (Number) Threshold value for comparison.

## Import

Import existing alert rules using the format `busi_group_id:alert_rule_id`:

```shell
terraform import nightingale_alert_rule.example 1:123
```
