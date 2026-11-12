resource "bronto_group" "platform" {
  name        = "Platform Team"
  description = "Owns core platform infrastructure"

  # Optional: reconciled via individual add/remove membership API calls,
  # not the group's own create/update endpoints.
  members = [
    {
      member_type = "USER"
      member_id   = bronto_user.alice.id
    },
    {
      member_type = "GROUP"
      member_id   = "02bd4c9a-e839-4fc1-b1b4-7490a1f354e7"
    }
  ]
}
