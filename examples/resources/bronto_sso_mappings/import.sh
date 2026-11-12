# This is a singleton resource - the id is always the constant
# "sso_mappings".
terraform import bronto_sso_mappings.this sso_mappings

# Alternatively, the Bronto API can generate a ready-to-paste
# bronto_sso_mappings block (including its own `import` block) for the
# org's existing mappings directly:
#   GET /sso/mappings/export?format=terraform
