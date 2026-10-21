# The workspace must already be connected through the Bronto UI's OAuth
# flow - look up its ID with the bronto_slack_workspaces data source.
data "bronto_slack_workspaces" "all" {}

resource "bronto_slack_integration" "api_alerts" {
  name         = "api_alerts"
  workspace_id = [for w in data.bronto_slack_workspaces.all.workspaces : w.id if w.name == "brontotest"][0]

  channels = [
    {
      name = "production-alerts"
      # include_graph_image defaults to true.
    },
    {
      name                = "usage-alerts"
      include_graph_image = false
    },
  ]
}
