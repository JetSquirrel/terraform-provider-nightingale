provider "nightingale" {
  endpoint = var.nightingale_endpoint
  token    = var.nightingale_token
}

# Notification rule for email alerts
resource "nightingale_notify_rule" "email_ops" {
  name           = "Email OPS Team"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}

# Critical alert rule: disk nearly full
resource "nightingale_alert_rule" "disk_critical" {
  busi_group_id    = var.busi_group_id
  name             = "Disk nearly full"
  datasource_type  = "prometheus"
  datasource_ids   = [var.datasource_id]
  severity         = 1
  notify_recovered = true

  queries = [{
    ref              = "A"
    promql           = "node_filesystem_avail_bytes{fstype!~\"tmpfs|squashfs\"} / node_filesystem_size_bytes * 100 < 10"
    duration_seconds = 600
  }]

  annotations = {
    summary     = "Disk nearly full on {{ $labels.ident }}"
    description = "Available disk space is below 10%."
  }

  append_tags = [
    "severity=critical",
    "team=ops",
    "managed_by=terraform",
  ]

  notify_rule_ids = [nightingale_notify_rule.email_ops.id]
}

# Warning alert rule: high CPU
resource "nightingale_alert_rule" "cpu_warning" {
  busi_group_id    = var.busi_group_id
  name             = "High CPU usage"
  datasource_type  = "prometheus"
  datasource_ids   = [var.datasource_id]
  severity         = 2
  notify_recovered = true

  queries = [{
    ref              = "A"
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100"
    duration_seconds = 300
  }]

  annotations = {
    summary     = "High CPU usage on {{ $labels.ident }}"
    description = "CPU usage has been above threshold for 5 minutes."
  }

  append_tags = [
    "severity=warning",
    "managed_by=terraform",
  ]

  notify_rule_ids = [nightingale_notify_rule.email_ops.id]
}

# Subscription for critical alerts to the ops team
resource "nightingale_alert_subscribe" "ops_critical" {
  busi_group_id = var.busi_group_id
  name          = "OPS Critical Subscription"
  disabled      = false

  rule_ids   = [nightingale_alert_rule.disk_critical.id]
  severities = [1]

  user_group_ids  = [1]
  notify_rule_ids = [nightingale_notify_rule.email_ops.id]
}
