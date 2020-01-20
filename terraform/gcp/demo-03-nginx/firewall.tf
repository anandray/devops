resource "google_compute_firewall" "allow_http" {
 name    = "nginx-firewall"
 network = "default"

 allow {
   protocol = "tcp"
   ports    = ["80"]
 }
 
 target_tags = ["allow-http"]

}
