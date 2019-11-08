#!/bin/bash

## Usage: ./create-snapshots-images.sh d5 --> will first create 
## a sinapshot named 'd5-date' and then use that snapshot to 
## create an image named 'd5-date'. The script will delete the
## snapshot it created once the image creation is complete.
## First run "screen -mS image" to enter screen mode and then run the script.

date=`date +%m%d%Y`
echo "date=`date +%m%d%Y`" > /tmp/snapshot-$1.sh
echo "date=`date +%m%d%Y`" > /tmp/image-$1.sh
echo "date=`date +%m%d%Y`" > /tmp/snapshot-delete-$1.sh

# Create snapshots to be used to create images
gcloud compute instances --project=us-west1-wlk list | grep RUNNING | grep $1 | awk '{print"gcloud beta compute --project=us-west1-wlk disks snapshot",$1,"--zone="$2,"--snapshot-names="$1"-\$date","--storage-location=us --verbosity info --format=json;"}' >> /tmp/snapshot-$1.sh && sh /tmp/snapshot-$1.sh

# Create images from snapshots created above
gcloud compute snapshots --project=us-west1-wlk  list | grep READY | grep $1-$date | awk '{print"gcloud compute --project=us-west1-wlk images create --source-snapshot="$1,$1,"--verbosity info --format=json;"}' >> /tmp/image-$1.sh && sh /tmp/image-$1.sh

# Delete the snapshots once the images are created
gcloud -q compute snapshots --project=us-west1-wlk list | grep READY | grep $1-$date | awk '{print"gcloud -q compute snapshots --project=us-west1-wlk delete",$1,"--verbosity info;"}' >> /tmp/snapshot-delete.sh && sh /tmp/snapshot-delete.sh
