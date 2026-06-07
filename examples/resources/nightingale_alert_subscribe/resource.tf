resource "nightingale_alert_subscribe" "example" {
  busi_group_id = 1
  name          = "OPS Critical Alerts"
  disabled      = false

  severities      = [1, 2]
  user_group_ids  = [1]
  notify_rule_ids = [1]
}
