resource "bronto_widget" "errors_by_service" {
  name        = "5xx errors by service"
  description = "This widget shows the 5xx errors by service"
  type        = "treemap"

  # A widget can include either sources or widget_ids, but not both. Sources
  # can reference a metric definition, a monitor, or an SLO.
  sources = [
    {
      type = "metric"
      id   = "41521f35-1b06-41f0-9cdc-9a938b5739d1"
    }
  ]

  # Optional: arbitrary widget-type-specific configuration.
  aux = jsonencode({
    color_name = "Purple"
    mode       = "sql"
  })
}

# A composite widget that groups other widgets instead of sources.
resource "bronto_widget" "overview_group" {
  name       = "Service overview"
  type       = "top-list"
  widget_ids = [bronto_widget.errors_by_service.id]
}
