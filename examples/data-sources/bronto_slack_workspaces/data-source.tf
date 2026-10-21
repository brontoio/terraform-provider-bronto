data "bronto_slack_workspaces" "all" {}

output "connected_slack_workspaces" {
  value = data.bronto_slack_workspaces.all.workspaces
}
