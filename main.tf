terraform {
  required_providers {
    bronto = {
      source = "hashicorp.com/io/bronto"
    }
  }
}

variable "bronto_api_key" {
  type      = string
  sensitive = true
}

provider "bronto" {
  endpoint = "https://api.eu.staging.bronto.io"
  api_key  = var.bronto_api_key
}

resource "bronto_user" "demo_import" {
  email        =  "gary+tfimporting3@bronto.io"
  first_name   =  "Demo"
  last_name    = "Complete"
  # This org has multi-org enabled: roles are managed via group/access-grant
  # assignment, not directly on the user, and the API now rejects (400) any
  # non-empty roles here. Empty list matches what the real UI sends too.
  roles        = []

  # NOTE: the backend silently no-ops clearing login_methods to empty (see
  # bronto-meta UserStore.updateUser -> DynamoExpressionBuilder.addExpression,
  # which skips the update expression entirely when the value is empty) - so
  # this must stay non-empty to match reality; setting it to [] here is a
  # known-broken, untestable state.
  login_methods = [
    {
      name    = "Password"
      enabled = true
    },
    {
      name    = "Google"
      enabled = false
    }
  ]

  active = true

  tags = {
    purpose = "tags-feature-verification"
  }
}

data "bronto_user" "demo_lookup" {
  email      = bronto_user.demo_import.email
  depends_on = [bronto_user.demo_import]
}

data "bronto_user" "demo_lookup_by_id" {
  id         = bronto_user.demo_import.id
  depends_on = [bronto_user.demo_import]
}

# The bronto_metric_definition / bronto_widget / bronto_dashboard demo
# (api_gateway_search_calls / api_gateway_search_activity) now lives in the
# separate `dogfood` repo, which dogfoods this provider as a real consumer
# would. The resources themselves weren't destroyed - just moved (via
# `terraform state rm` here + `terraform import` there) to avoid two states
# fighting over the same remote objects.
