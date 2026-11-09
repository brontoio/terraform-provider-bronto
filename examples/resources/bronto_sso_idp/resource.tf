resource "bronto_sso_idp" "okta" {
  name          = "Okta"
  provider_type = "SAML"
  identifiers   = ["bronto.io"]
  meta_cert     = file("${path.module}/okta-metadata.xml")

  attribute_mappings = [
    {
      provider_attr = "givenname"
      user_attr     = "name"
    }
  ]
}

# OIDC providers use issuer_url/client_id/client_secret instead of
# meta_cert.
resource "bronto_sso_idp" "google" {
  name          = "Google"
  provider_type = "OIDC"
  issuer_url    = "https://accounts.google.com"
  client_id     = var.google_oidc_client_id
  client_secret = var.google_oidc_client_secret

  attribute_mappings = [
    {
      provider_attr = "email"
      user_attr     = "email"
    }
  ]
}
