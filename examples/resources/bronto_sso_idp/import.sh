# SSO identity providers are imported by name - this resource has no
# separate generated id.
terraform import bronto_sso_idp.okta Okta

# Alternatively, the Bronto API can generate a ready-to-paste bronto_sso_idp
# block (including its own `import` block) for an existing identity
# provider directly. client_secret is never returned by the API, so an
# OIDC export will need it filled in by hand afterward:
#   GET /sso/idps/{idp_name}/export?format=terraform
