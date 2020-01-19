resource "google_compute_firewall" "tfer--default-002D-allow-002D-icmp" {
  allow {
    protocol = "icmp"
  }

  description    = "Allow ICMP from anywhere"
  direction      = "INGRESS"
  disabled       = "false"
  enable_logging = "false"
  name           = "default-allow-icmp"
  network        = "${data.terraform_remote_state.networks.outputs.google_compute_network_tfer--default_self_link}"
  priority       = "65534"
  project        = "devops-112019"
  source_ranges  = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "tfer--default-002D-allow-002D-internal" {
  allow {
    ports    = ["0-65535"]
    protocol = "udp"
  }

  allow {
    protocol = "icmp"
  }

  allow {
    ports    = ["0-65535"]
    protocol = "tcp"
  }

  description    = "Allow internal traffic on the default network"
  direction      = "INGRESS"
  disabled       = "false"
  enable_logging = "false"
  name           = "default-allow-internal"
  network        = "${data.terraform_remote_state.networks.outputs.google_compute_network_tfer--default_self_link}"
  priority       = "65534"
  project        = "devops-112019"
  source_ranges  = ["10.128.0.0/9"]
}

resource "google_compute_firewall" "tfer--default-002D-allow-002D-rdp" {
  allow {
    ports    = ["3389"]
    protocol = "tcp"
  }

  description    = "Allow RDP from anywhere"
  direction      = "INGRESS"
  disabled       = "false"
  enable_logging = "false"
  name           = "default-allow-rdp"
  network        = "${data.terraform_remote_state.networks.outputs.google_compute_network_tfer--default_self_link}"
  priority       = "65534"
  project        = "devops-112019"
  source_ranges  = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "tfer--default-002D-allow-002D-ssh" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }

  description    = "Allow SSH from anywhere"
  direction      = "INGRESS"
  disabled       = "false"
  enable_logging = "false"
  name           = "default-allow-ssh"
  network        = "${data.terraform_remote_state.networks.outputs.google_compute_network_tfer--default_self_link}"
  priority       = "65534"
  project        = "devops-112019"
  source_ranges  = ["0.0.0.0/0"]
}
