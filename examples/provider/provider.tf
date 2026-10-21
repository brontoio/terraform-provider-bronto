terraform {
  required_providers {
    bronto = {
      source = "brontoio/bronto"
    }
  }
}

provider "bronto" {
  # endpoint and api_key can also be supplied via the BRONTO_ENDPOINT and
  # BRONTO_API_KEY environment variables.
  endpoint = "https://api.eu.bronto.io"
  api_key  = var.bronto_api_key
}

# Managing multiple orgs (e.g. a root org and a child org, each with their
# own API key) from one configuration: declare multiple aliased instances
# and point each resource at the right one with provider = bronto.<alias>.
provider "bronto" {
  alias    = "root"
  endpoint = var.bronto_endpoint
  api_key  = var.root_org_api_key
}

provider "bronto" {
  alias    = "child"
  endpoint = var.bronto_endpoint
  api_key  = var.child_org_api_key
}

resource "bronto_user" "example" {
  provider = bronto.root
  # ...
}
