// Configure the Google Cloud provider
provider "google" {
# credentials = "${file("Terraform.json")}"
 credentials = file(var.credentials_file_path)
 project     = "devops-112019"
 region      = "us-west1"
}

// Terraform plugin for creating random ids
resource "random_id" "instance_id" {
 byte_length = 4
}

// A single Google Cloud Engine instance
resource "google_compute_instance" "default" {
 name         = "devops-${random_id.instance_id.hex}"
 machine_type = "f1-micro"
 zone         = "us-west1-a"

 boot_disk {
   initialize_params {
     image = "debian-cloud/debian-9"
   }
 }

// Make sure flask is installed on all new instances for later steps
  metadata_startup_script = "sudo apt-get update; sudo apt-get install -y build-essential telnet curl nginx; sudo systemctl start nginx"

 network_interface {
   network = "default"

   access_config {
     // Include this section to give the VM an external ip address
   }
 }
 metadata = {
   ssh-keys = "anand:${file("~/.ssh/id_rsa.pub")}"
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
