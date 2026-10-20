resource "bronto_widget" "errors_by_service" {
  name = "5xx errors by service"
  type = "treemap"
  sources = [
    {
      type = "metric"
      id   = "41521f35-1b06-41f0-9cdc-9a938b5739d1"
    }
  ]
}

resource "bronto_dashboard" "application_performance" {
  name       = "Application Performance Dashboard"
  widget_ids = [bronto_widget.errors_by_service.id]

  # Optional: metrics (LBMs) associated with the dashboard as a whole.
  layout_metric_ids = ["41521f35-1b06-41f0-9cdc-9a938b5739d1"]

  # Optional: each key change costs a separate API call (there's no bulk tag
  # endpoint), so this is best set once rather than churned frequently.
  tags = {
    team        = "observability"
    environment = "production"
  }
}
