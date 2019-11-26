#!/bin/bash
gsutil cp gs://startup_scripts_us/scripts/denyhosts/allowed-hosts /var/lib/denyhosts/allowed-hosts
service denyhosts restart
