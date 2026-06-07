resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = 1
  name          = "OPS Critical"
  disabled      = false

  rule_ids   = [10, 11]
  severities = [1, 2]

  user_group_ids  = [5]
  notify_rule_ids = [3]

  tags        = "env=prod"
  busi_groups = "ops"
}
