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
