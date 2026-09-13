# Look up a user by id.
data "bronto_user" "by_id" {
  id = "41521f35-1b06-41f0-9cdc-9a938b5739d1"
}

# Or by email.
data "bronto_user" "by_email" {
  email = "jane.doe@example.com"
}

output "jane_doe_roles" {
  value = data.bronto_user.by_email.roles
}
