resource "bronto_user" "example" {
  first_name = "Jane"
  last_name  = "Doe"
  email      = "jane.doe@example.com"
  roles      = ["Standard"]

  # Optional: disable individual login methods for this user. Any method
  # not listed here uses the organization-wide default.
  login_methods = [
    {
      name    = "Password"
      enabled = false
    },
  ]
}

# Deactivate a user without deleting them, keeping their history and access
# grants intact. Setting `active` back to true reactivates them.
resource "bronto_user" "on_leave" {
  first_name = "Sam"
  last_name  = "Taylor"
  email      = "sam.taylor@example.com"
  roles      = ["ReadOnly"]
  active     = false
}
