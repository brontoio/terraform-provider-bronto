# Look up by ID.
data "bronto_dataset" "by_id" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}

# Or by collection and dataset name together - both must be set, and
# exactly one of id, or collection and dataset, is required.
data "bronto_dataset" "firewall_logs" {
  collection = "demo"
  dataset    = "firewall"
}

output "firewall_dataset_id" {
  value = data.bronto_dataset.firewall_logs.id
}
