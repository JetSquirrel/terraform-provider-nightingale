provider "nightingale" {
  endpoint = var.nightingale_endpoint
  token    = var.nightingale_token
}

# Notification rule
resource "nightingale_notify_rule" "ops" {
  name           = "OPS Team Notifications"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}

# Alert rule: High CPU
resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = var.busi_group_id
  name            = "High CPU Usage"
  datasource_type = "prometheus"
  datasource_ids  = [var.datasource_id]
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

  append_tags     = ["managed_by=terraform"]
  notify_rule_ids = [nightingale_notify_rule.ops.id]
}

# Alert rule: Disk Full
resource "nightingale_alert_rule" "disk_full" {
  busi_group_id   = var.busi_group_id
  name            = "Disk Nearly Full"
  datasource_type = "prometheus"
  datasource_ids  = [var.datasource_id]
  severity        = 1

  queries = [{
    ref              = "A"
    promql           = "node_filesystem_avail_bytes / node_filesystem_size_bytes * 100 < 10"
    duration_seconds = 600
  }]

  annotations = {
    summary     = "Disk nearly full on {{ $labels.ident }}"
    description = "Available disk space is below 10%."
  }

  append_tags     = ["managed_by=terraform", "severity=critical"]
  notify_rule_ids = [nightingale_notify_rule.ops.id]
}

# Alert subscription
resource "nightingale_alert_subscribe" "critical" {
  busi_group_id = var.busi_group_id
  name          = "Critical Alerts Subscription"
  disabled      = false

  rule_ids        = [nightingale_alert_rule.disk_full.id]
  severities      = [1]
  user_group_ids  = [1]
  notify_rule_ids = [nightingale_notify_rule.ops.id]
}
