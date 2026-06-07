---
page_title: "nightingale_notify_rule Resource - Nightingale"
description: |-
  Manages a Nightingale notification rule.
---

# nightingale_notify_rule

Manages a Nightingale notification rule. Notification rules define how alerts are delivered to users through various channels.

## Example Usage

### Basic Notification Rule

```terraform
resource "nightingale_notify_rule" "ops_email" {
  name           = "OPS Team Email"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}
```

### Multiple Channels

```terraform
resource "nightingale_notify_rule" "multi_channel" {
  name           = "Multi-Channel Alerts"
  enable         = true
  user_group_ids = [1, 2]

  notify_configs = [
    {
      channel_id  = 1  # Email
      template_id = 1
    },
    {
      channel_id  = 2  # Webhook
      template_id = 2
      params = {
        url = "https://hooks.example.com/alerts"
      }
    }
  ]
}
```

## Schema

### Required

- `name` (String) Notification rule name.
- `notify_configs` (Block List, Min: 1) Notification channel configurations. See [Nested Schema](#nestedblock--notify_configs) below.
- `user_group_ids` (Set of Number) User group IDs to notify.

### Optional

- `enable` (Boolean) Whether the rule is enabled. Default: `true`.

### Read-Only

- `create_at` (Number) Creation timestamp.
- `create_by` (String) Creator username.
- `id` (String) Notification rule ID.
- `update_at` (Number) Last update timestamp.
- `update_by` (String) Last updater username.

<a id="nestedblock--notify_configs"></a>
### Nested Schema for `notify_configs`

Required:

- `channel_id` (Number) Notification channel ID.

Optional:

- `params` (Map of String) Channel-specific parameters (e.g., webhook URL, secret).
- `template_id` (Number) Message template ID.

## Import

Import existing notification rules using the rule ID:

```shell
terraform import nightingale_notify_rule.example 456
```
