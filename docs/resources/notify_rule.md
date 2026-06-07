# nightingale_notify_rule

Manages a Nightingale notification rule.

## Example Usage

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

## Schema

### Required

- `name` (String) Notification rule name.
- `notify_configs` (Attributes List) Notification channel configurations. (see [below for nested schema](#nestedatt--notify_configs))
- `user_group_ids` (Set of Number) User group IDs associated with this rule.

### Optional

- `enable` (Boolean) Whether the notification rule is enabled. Default is `true`.

### Read-Only

- `create_at` (Number) Remote creation timestamp.
- `create_by` (String) Remote creator.
- `id` (String) Notification rule ID.
- `update_at` (Number) Remote update timestamp.
- `update_by` (String) Remote updater.

<a id="nestedatt--notify_configs"></a>
### Nested Schema for `notify_configs`

Required:

- `channel_id` (Number) Notification channel ID.

Optional:

- `params` (Map of String) Custom parameters for the notification channel.
- `template_id` (Number) Message template ID.

## Import

Import is supported using the following syntax:

```shell
terraform import nightingale_notify_rule.example <notify_rule_id>
```
