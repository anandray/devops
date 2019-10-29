#!/bin/bash
/usr/bin/geoipupdate -v
gsutil -m cp -r /usr/share/GeoIP/* gs://startup_scripts_us/scripts/GeoIP/
