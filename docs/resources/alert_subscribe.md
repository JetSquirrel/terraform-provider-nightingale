---
page_title: "nightingale_alert_subscribe Resource - Nightingale"
description: |-
  Manages a Nightingale alert subscription rule.
---

# nightingale_alert_subscribe

Manages a Nightingale alert subscription. Subscriptions allow users or teams to receive notifications for specific alerts based on severity, tags, or other criteria.

## Example Usage

### Basic Subscription

```terraform
resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = 1
  name          = "OPS Critical Alerts"
  disabled      = false

  severities      = [1, 2]
  user_group_ids  = [1]
  notify_rule_ids = [1]
}
```

### Subscription with Filters

```terraform
resource "nightingale_alert_subscribe" "prod_alerts" {
  busi_group_id = 1
  name          = "Production Alerts"
  disabled      = false

  rule_ids        = [10, 11, 12]
  severities      = [1]
  datasource_ids  = [1]
  
  tags        = "env=prod"
  busi_groups = "production"

  user_group_ids  = [5]
  notify_rule_ids = [3]
}
```

### Subscription Linked to Alert Rules

```terraform
resource "nightingale_alert_subscribe" "disk_alerts" {
  busi_group_id = 1
  name          = "Disk Alert Subscription"

  rule_ids        = [nightingale_alert_rule.disk_full.id]
  severities      = [1]
  user_group_ids  = [1]
  notify_rule_ids = [nightingale_notify_rule.ops.id]
}
```

## Schema

### Required

- `busi_group_id` (Number) Business group ID. Changing this forces a new resource.
- `name` (String) Subscription rule name.

### Optional

- `busi_groups` (String) Business group filter expression.
- `datasource_ids` (Set of Number) Filter alerts by datasource IDs.
- `disabled` (Boolean) Whether the subscription is disabled. Default: `false`.
- `notify_rule_ids` (Set of Number) Notification rule IDs to use.
- `notify_version` (Number) Notify version. Default: `1`.
- `rule_ids` (Set of Number) Subscribe to specific alert rule IDs.
- `severities` (Set of Number) Filter by severity levels (1=Critical, 2=Warning, 3=Info).
- `tags` (String) Tag filter expression (e.g., `env=prod`).
- `user_group_ids` (Set of Number) User group IDs to notify.

### Read-Only

- `create_at` (Number) Creation timestamp.
- `create_by` (String) Creator username.
- `id` (String) Subscription rule ID.
- `update_at` (Number) Last update timestamp.
- `update_by` (String) Last updater username.

## Import

Import existing subscriptions using the format `busi_group_id:subscribe_id`:

```shell
terraform import nightingale_alert_subscribe.example 1:789
```
