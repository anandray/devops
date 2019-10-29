#!/bin/bash
date=`date +%m%d%Y`
echo "date=`date +%m%d%Y`" > /tmp/snapshot-$1.sh
echo "date=`date +%m%d%Y`" > /tmp/image-$1.sh
echo "date=`date +%m%d%Y`" > /tmp/snapshot-delete-$1.sh

#gcloud beta compute --project=crosschannel-1307 disks snapshot d1 --zone=us-central1-c --snapshot-names=d1-$date --storage-location=us
#gcloud compute --project=crosschannel-1307 images create --source-snapshot=d1-$date d1-$date
#gcloud -q compute snapshots --project=crosschannel-1307 delete d1-$date

# Create snapshots to be used to create images
gcloud compute instances --project=us-west1-wlk list | grep RUNNING | grep $1 | awk '{print"gcloud beta compute --project=us-west1-wlk disks snapshot",$1,"--zone="$2,"--snapshot-names="$1"-\$date","--storage-location=us --verbosity info --format=json;"}' >> /tmp/snapshot-$1.sh && sh /tmp/snapshot-$1.sh

# Create images from snapshots created above
gcloud compute snapshots --project=us-west1-wlk  list | grep READY | grep $1-$date | awk '{print"gcloud compute --project=us-west1-wlk images create --source-snapshot="$1,$1,"--verbosity info --format=json;"}' >> /tmp/image-$1.sh && sh /tmp/image-$1.sh

# Delete the snapshots once the images are created
gcloud -q compute snapshots --project=us-west1-wlk list | grep READY | grep $1-$date | awk '{print"gcloud -q compute snapshots --project=us-west1-wlk delete",$1,"--verbosity info;"}' >> /tmp/snapshot-delete.sh && sh /tmp/snapshot-delete.sh
