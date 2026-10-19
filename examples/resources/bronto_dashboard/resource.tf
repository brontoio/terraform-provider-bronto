resource "bronto_widget" "errors_by_service" {
  name = "5xx errors by service"
  type = "treemap"
  # Metrics (LBMs) aren't managed by this provider yet; substitute a real metric ID.
  metric_ids = ["41521f35-1b06-41f0-9cdc-9a938b5739d1"]
}

resource "bronto_dashboard" "application_performance" {
  name       = "Application Performance Dashboard"
  widget_ids = [bronto_widget.errors_by_service.id]

  # Optional: metrics (LBMs) associated with the dashboard as a whole.
  layout_metric_ids = ["41521f35-1b06-41f0-9cdc-9a938b5739d1"]
}
