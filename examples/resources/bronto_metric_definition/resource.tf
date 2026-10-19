resource "bronto_metric_definition" "api_response_time" {
  name        = "api_response_time"
  description = "Tracks response time for API endpoints in production"

  queries = [
    {
      name      = "production_api_response_time"
      select    = ["response_time_ms"]
      from_expr = "environment = 'production' AND service = 'api-gateway'"
      where     = "endpoint LIKE '/api/v1/%' AND response_time_ms > 0"
      groups    = ["endpoint", "method"]

      aggregation = {
        time      = "avg"
        reduce_to = "avg"
      }
    }
  ]

  # Optional: combine the queries above into a derived value.
  formulas = [
    {
      name       = "response_time_seconds"
      expression = "production_api_response_time / 1000"
    }
  ]
}

# The resulting metric definition ID can be referenced from a widget.
resource "bronto_widget" "api_response_time" {
  name       = "API response time"
  type       = "line"
  metric_ids = [bronto_metric_definition.api_response_time.id]
}
