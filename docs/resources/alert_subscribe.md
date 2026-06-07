# nightingale_alert_subscribe

Manages a Nightingale alert subscription rule.

## Example Usage

```terraform
resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = 1
  name          = "OPS Critical"
  disabled      = false

  rule_ids = [10, 11]
  severities = [1, 2]

  user_group_ids  = [5]
  notify_rule_ids = [3]

  tags       = "env=prod"
  busi_groups = "ops"
}
```

## Schema

### Required

- `busi_group_id` (Number) Nightingale business group ID. Changing this forces a new resource.
- `name` (String) Subscription rule name.

### Optional

- `busi_groups` (String) Business group filter expression.
- `datasource_ids` (Set of Number) Datasource IDs to filter alerts.
- `disabled` (Boolean) Whether the subscription is disabled. Default is `false`.
- `notify_rule_ids` (Set of Number) Notification rule IDs to use.
- `notify_version` (Number) Notify version (1 for new notify rules). Default is `1`.
- `rule_ids` (Set of Number) Alert rule IDs to subscribe to.
- `severities` (Set of Number) Severities to match.
- `tags` (String) Tag filter expression.
- `user_group_ids` (Set of Number) User group IDs to notify.

### Read-Only

- `create_at` (Number) Remote creation timestamp.
- `create_by` (String) Remote creator.
- `id` (String) Alert subscription rule ID.
- `update_at` (Number) Remote update timestamp.
- `update_by` (String) Remote updater.

## Import

Import is supported using the following syntax:

```shell
terraform import nightingale_alert_subscribe.example <busi_group_id>:<alert_subscribe_id>
```
