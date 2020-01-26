
resource "google_compute_firewall"  "allow_http" {
    name = "allow-http"
    network = "default"

    allow {
        protocol = "tcp"
        ports = ["80"]
    }

    target_tags = ["allow-http"]

}

resource "google_compute_firewall"  "allow_https" {
    name = "allow-https"
    network = "default"

    allow {
        protocol = "tcp"
        ports = ["443"]
    }

    target_tags = ["allow-https"]

}

resource "google_compute_firewall"  "allow_one" {
    name = "allow-one"
    network = "default"

    allow {
        protocol = "tcp"
        ports = ["8080"]
    }

    target_tags = ["allow-one"]

}

resource "google_compute_firewall"  "allow_two" {
    name = "allow-two"
    network = "default"

    allow {
        protocol = "tcp"
        ports = ["999"]
    }

    target_tags = ["allow-two"]

}