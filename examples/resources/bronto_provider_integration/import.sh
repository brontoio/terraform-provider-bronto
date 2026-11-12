# Provider integrations can be imported by their unique ID. The secret
# can't be imported (the API never returns it) and must be set in config
# afterward if you want Terraform to manage it going forward.
terraform import bronto_provider_integration.pagerduty 550e8400-e29b-41d4-a716-446655440000

# Alternatively, the Bronto API can generate a ready-to-paste
# bronto_provider_integration block (including its own `import` block) for
# an existing integration directly. secret is never returned by the API,
# so it will still need to be filled in by hand afterward:
#   GET /integrations/providers/{provider_id}/export?format=terraform
