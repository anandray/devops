#!/bin/bash
date=`date +%m%d%Y`
echo "date=`date +%m%d%Y`" > /tmp/snapshot.sh
echo "date=`date +%m%d%Y`" > /tmp/image.sh
echo "date=`date +%m%d%Y`" > /tmp/snapshot-delete.sh

#gcloud beta compute --project=crosschannel-1307 disks snapshot d1 --zone=us-central1-c --snapshot-names=d1-$date --storage-location=us
#gcloud compute --project=crosschannel-1307 images create --source-snapshot=d1-$date d1-$date
#gcloud -q compute snapshots --project=crosschannel-1307 delete d1-$date

# Create snapshots to be used to create images
gcloud compute instances --project=crosschannel-1307 list | grep RUNNING | awk '{print"gcloud beta compute --project=crosschannel-1307 disks snapshot",$1,"--zone="$2,"--snapshot-names="$1"-\$date","--storage-location=us --verbosity info;"}' >> /tmp/snapshot.sh && sh /tmp/snapshot.sh

# Create images from snapshots created above
gcloud compute snapshots --project=crosschannel-1307  list | grep READY | awk '{print"gcloud compute --project=crosschannel-1307 images create --source-snapshot="$1,$1"--verbosity info;"}' >> /tmp/image.sh && sh /tmp/image.sh

# Delete the snapshots once the images are created
gcloud -q compute snapshots --project=crosschannel-1307 list | grep READY | awk '{print"gcloud -q compute snapshots --project=crosschannel-1307 delete",$1,"--verbosity info;"}' >> /tmp/snapshot-delete.sh && sh /tmp/snapshot-delete.sh
