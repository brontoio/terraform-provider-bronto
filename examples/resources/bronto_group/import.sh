# Groups can be imported by their unique ID.
terraform import bronto_group.platform 02bd4c9a-e839-4fc1-b1b4-7490a1f354e7

# Alternatively, the Bronto API can generate a ready-to-paste bronto_group
# block (including its own `import` block) for an existing group directly:
#   GET /groups/{groupId}/export?format=terraform
