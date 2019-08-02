#!/bin/bash
#/usr/bin/geoipupdate -v
gsutil -m cp -r gs://startup_scripts_us/scripts/GeoIP/* /usr/share/GeoIP/
