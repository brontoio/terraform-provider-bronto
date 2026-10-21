resource "bronto_webhook_integration" "on_call" {
  name = "on-call-pager"
  url  = "https://webhook.site/cfee291c-4fcc-464a-aabc-3f9653295db3"

  # Optional: custom_headers is marked sensitive since this is the only way
  # to supply auth (e.g. an API key or bearer token) to the webhook.
  custom_headers = {
    authorization = "Bearer xxxxx"
  }

  # Optional: the request body template, with $variable placeholders
  # resolved from the triggering alert.
  payload = jsonencode({
    name    = "$title"
    message = "$description"
    details = { id = "$alert_id" }
  })
}
