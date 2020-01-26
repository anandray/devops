module "gke" {
  source                     = "terraform-google-modules/kubernetes-engine/google"
  project_id                 = var.project
  name                       = "gke-k8s-012020"
  region                     = var.region
  zones                      = ["us-west1-a", "us-west1-b", "us-west1-c"]
#  network                    = "vpc-01"
#  subnetwork                 = "us-west1-01"
#  ip_range_pods              = "us-west1-01-gke-01-pods"
#  ip_range_services          = "us-west1-01-gke-01-services"
  network                    = "default"
  subnetwork                 = "default"
  ip_range_pods              = ""
  ip_range_services          = ""
  http_load_balancing        = false
  horizontal_pod_autoscaling = true
  #network_policy             = true
  network_policy             = false

  node_pools = [
    {
      name               = "default-node-pool"
#      machine_type       = "n1-standard-1"
      machine_type       = "f1-micro"
      min_count          = 1
      max_count          = 2
      local_ssd_count    = 0
      disk_size_gb       = 10
      disk_type          = "pd-standard"
      image_type         = "COS"
      auto_repair        = true
      auto_upgrade       = true
#      service_account    = "project-service-account@devops-112019.iam.gserviceaccount.com"
      service_account    = "terraform@devops-112019.iam.gserviceaccount.com"
      preemptible        = false
      initial_node_count = 1
    },
  ]

  node_pools_oauth_scopes = {
    all = []

    default-node-pool = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  node_pools_labels = {
    all = {}

    default-node-pool = {
      default-node-pool = true
    }
  }

  node_pools_metadata = {
    all = {}

    default-node-pool = {
      node-pool-metadata-custom-value = "my-node-pool"
    }
  }

#  node_pools_taints = {
#    all = []
#
#    default-node-pool = [
#      {
#        key    = "default-node-pool"
#        value  = true
#        effect = "PREFER_NO_SCHEDULE"
#      },
#    ]
#  }

  node_pools_tags = {
    all = []

    default-node-pool = [
      "default-node-pool",
    ]
  }
}
