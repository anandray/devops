#!/bin/bash
/usr/local/share/google/google-cloud-sdk/bin/gsutil -m rsync -e -C -p -R /var/www/vhosts/g2.mdotm.co/httpdocs/ gs://crosschannel_cdn/
