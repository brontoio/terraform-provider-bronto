terraform {
  required_providers {
    bronto = {
      source = "hashicorp.com/io/bronto"
    }
  }
}

provider "bronto" {
  region = "staging"
  api_key = "xxxx"
}

resource "bronto_user" "demo_import" {
  email        =  "gary+tfimporting@brontobytes.io"
  first_name   =  "Demo"
  last_name    = "Complete"
  roles        = ["Admin"]
}
