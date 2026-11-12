# This is a singleton: declare exactly one bronto_sso_mappings resource per
# org. It manages the organization's whole list of SSO mappings at once.
resource "bronto_sso_mappings" "this" {
  mappings = [
    {
      provider_key   = "member_of"
      provider_value = "Developers"
      assign_attr    = "Standard"
    },
    {
      provider_key   = "member_of"
      provider_value = "Admins"
      assign_attr    = "Admin"
    }
  ]
}
