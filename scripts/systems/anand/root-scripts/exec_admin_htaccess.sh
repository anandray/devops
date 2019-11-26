#!/bin/bash
#while inotifywait -q -e modify /root/scripts/admin_htaccess >/dev/null; do
#    echo "filename is changed"
    /usr/local/share/google/google-cloud-sdk/bin/gcloud compute instances list | grep www60- | grep -v chat | awk '{print"rsync -avz /root/scripts/admin_htaccess",$1":/var/www/vhosts/mdotm.com/httpdocs/.htaccess 2>&1 &"}' > /root/scripts/rsync_admin_htaccess.sh && sh /root/scripts/rsync_admin_htaccess.sh
    # do whatever else you need to do
#done
