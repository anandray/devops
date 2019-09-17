#!/bin/bash
gsutil cp gs://startup_scripts_us/scripts/dataproc/hosts_sl /home/anand;
sudo cat /home/anand/hosts_sl >> /etc/hosts
