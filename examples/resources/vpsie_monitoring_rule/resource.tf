resource "vpsie_monitoring_rule" "example" {
  rule_name = "high-cpu-alert"
  frequency = "1"
  status    = "1"

  rule {
    metric_type    = "cpu"
    condition      = "greater_than"
    threshold      = "80"
    threshold_type = "percentage"
    period         = "5"

    action {
      action_name = "send_alert"
      action_key  = "send_alert"
      email       = "admin@example.com"
    }
  }
}
