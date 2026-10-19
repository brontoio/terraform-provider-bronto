resource "bronto_widget" "errors_by_service" {
  name        = "5xx errors by service"
  description = "This widget shows the 5xx errors by service"
  type        = "treemap"
  # Metrics (LBMs) aren't managed by this provider yet; substitute a real metric ID.
  metric_ids = ["41521f35-1b06-41f0-9cdc-9a938b5739d1"]

  # Optional: arbitrary widget-type-specific configuration.
  aux = jsonencode({
    color_name = "Purple"
    mode       = "sql"
  })
}

# A composite widget that groups other widgets instead of metrics. A widget
# can reference metrics or other widgets, but not both.
resource "bronto_widget" "overview_group" {
  name       = "Service overview"
  type       = "top-list"
  widget_ids = [bronto_widget.errors_by_service.id]
}
