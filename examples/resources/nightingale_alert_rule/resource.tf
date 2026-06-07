resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = var.busi_group_id
  name            = "High CPU usage"
  datasource_type = "prometheus"
  datasource_ids  = [1]
  severity        = 2

  queries = [{
    ref              = "A"
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100"
    duration_seconds = 300
  }]

  annotations = {
    summary     = "High CPU usage on {{ $labels.ident }}"
    description = "CPU usage has been high for 5 minutes."
  }

  append_tags = [
    "managed_by=terraform",
  ]
}

resource "nightingale_alert_rule" "memory_threshold" {
  busi_group_id    = var.busi_group_id
  name             = "Memory usage threshold"
  datasource_type  = "prometheus"
  datasource_ids   = [1]
  severity         = 3
  disabled         = false
  notify_recovered = true

  queries = [{
    ref                = "A"
    promql             = "node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100 < 20"
    duration_seconds   = 180
    comparison_operator = "<"
    threshold          = 20
  }]

  annotations = {
    summary = "Memory low on {{ $labels.ident }}"
  }

  notify_rule_ids = [1]
}

resource "nightingale_alert_rule" "disk_full" {
  busi_group_id   = var.busi_group_id
  name            = "Disk nearly full"
  datasource_type = "prometheus"
  datasource_ids  = [1]
  severity        = 1

  queries = [{
    ref              = "A"
    promql           = "node_filesystem_avail_bytes{fstype!~\"tmpfs|squashfs\"} / node_filesystem_size_bytes * 100 < 10"
    duration_seconds = 600
  }]

  append_tags = ["severity=critical", "team=ops"]

  # Escape hatch for Nightingale-version-specific fields.
  extra_json = jsonencode({
    enable_stime = "00:00"
    enable_etime = "23:59"
  })
}
