#!/bin/bash
/usr/local/bin/gcloud compute instances list | grep www | grep RUNNING | grep -v www6001 | awk '{print"scp /etc/hosts",$1":/etc/ 2>&1 &"}' > /root/scripts/scp_hosts.sh;
sh /root/scripts/scp_hosts.sh
