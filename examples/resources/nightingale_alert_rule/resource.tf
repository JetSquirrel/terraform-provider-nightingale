resource "nightingale_alert_rule" "example" {
  busi_group_id   = 1
  name            = "High CPU Usage"
  datasource_type = "prometheus"
  datasource_ids  = [1]
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

  append_tags = ["managed_by=terraform"]
}
