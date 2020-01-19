resource "google_compute_network" "tfer--default" {
  auto_create_subnetworks         = "true"
  delete_default_routes_on_create = "false"
  description                     = "Default network for the project"
  name                            = "default"
  project                         = "devops-112019"
  routing_mode                    = "REGIONAL"
}
