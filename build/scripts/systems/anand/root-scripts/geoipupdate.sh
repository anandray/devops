#!/bin/bash
sudo gsutil cp gs://startup_scripts_us/scripts/GeoIP/GeoIP.conf /etc/GeoIP.conf
/usr/bin/geoipupdate -v
