resource "bronto_metric_definition" "error_rate" {
  name = "api_error_rate"
  queries = [
    {
      name   = "errors"
      select = ["count(*)"]
      where  = "status_code >= 500"
      from   = ["41521f35-1b06-41f0-9cdc-9a938b5739d1"]
      aggregation = {
        time      = "count"
        reduce_to = "avg"
      }
    }
  ]
}

# A monitor always references an existing bronto_metric_definition by id - it can't embed queries/formulas of its own.
resource "bronto_monitor" "error_rate" {
  name                = "High error rate"
  description         = "Error rate is above threshold over the last 20 minutes"
  metric_id           = bronto_metric_definition.error_rate.id
  comparison_operator = "ABOVE"
  threshold           = 50
  warning_threshold   = 25
  window              = "Last 20 minutes"

  actions = [
    {
      type  = "EMAIL"
      email = "oncall@example.com"
    }
  ]
}

# An anomaly detection monitor, comparing against a trailing baseline instead of a fixed threshold.
resource "bronto_monitor" "error_rate_anomaly" {
  name                = "Unusual error rate"
  metric_id           = bronto_metric_definition.error_rate.id
  monitor_type        = "ANOMALY_DETECTION"
  comparison_operator = "OUTSIDE"
  threshold           = 3
  evaluation_window   = "7 days"
  window              = "Last 20 minutes"

  actions = [
    {
      type  = "EMAIL"
      email = "oncall@example.com"
    }
  ]
}
