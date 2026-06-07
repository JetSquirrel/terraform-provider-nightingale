resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = var.busi_group_id
  name          = "OPS Critical"
  disabled      = false

  rule_ids   = [10, 11]
  severities = [1, 2]

  user_group_ids  = [5]
  notify_rule_ids = [3]

  tags        = "env=prod"
  busi_groups = "ops"
}

resource "nightingale_alert_subscribe" "all_severity" {
  busi_group_id = var.busi_group_id
  name          = "All severity alerts"
  disabled      = false

  severities = [1, 2, 3]

  user_group_ids  = [1]
  notify_rule_ids = [1]
}
