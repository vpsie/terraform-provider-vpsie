resource "vpsie_monitoring_rule" "example" {
  rule_name      = "high-cpu-alert"
  metric_type    = "cpu"
  condition      = "above"
  threshold_type = "percent"
  threshold      = "80"
  period         = "5m"
  frequency      = "1m"
  email          = "admin@example.com"
}
