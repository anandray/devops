data "google_compute_zones" "available" {}

resource "google_compute_instance" "default" {
  project      = google_project.project.project_id
  zone         = data.google_compute_zones.available.names[0]
  name         = "tf-compute-1"
  machine_type = "f1-micro"

  boot_disk {
    initialize_params {
      image = "ubuntu-1604-xenial-v20170328"
    }
  }

  network_interface {
    network       = "default"
    access_config {}
  }

   metadata = {
   ssh-keys = "anand:${file("~/.ssh/id_rsa.pub")}"
 }

     provisioner "file" {
    source      = "script.sh"
    destination = "/tmp/script.sh"
  }
  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/script.sh",
      "sudo /tmp/script.sh",
    ]
  }
}

resource "google_compute_firewall" "default" {
 name    = "nginx-firewall"
 network = "default"

 allow {
   protocol = "tcp"
   ports    = ["80"]
 }
}

// A variable for extracting the external ip of the instance
output "ip" {
 value = "${google_compute_instance.default.network_interface.0.access_config.0.nat_ip}"
}

output "instance_id" {
  value = google_compute_instance.default.self_link
}
