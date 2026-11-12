resource "bronto_provider_integration" "pagerduty" {
  name          = "PagerDuty"
  provider_type = "pagerduty"
  url           = "https://events.pagerduty.com/v2/enqueue"
  auth_type     = "api_key"

  # Required on create. Write-only: the API never returns it, so this is
  # never refreshed from the API - only sent on create, or again if you
  # change it to rotate the stored secret.
  secret = var.pagerduty_integration_key

  tags = {
    team = "observability"
  }
}
