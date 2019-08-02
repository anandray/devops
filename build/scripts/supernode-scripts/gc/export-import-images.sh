cloudservices_service_account=$(gcloud projects get-iam-policy wolk-us-west | grep cloudservices.gserviceaccount.com | cut -d ":" -f2)

gcloud projects add-iam-policy-binding wolk-1307 --member serviceAccount:364459736445@cloudbuild.gserviceaccount.com --role roles/editor
#gcloud projects add-iam-policy-binding wolk-1307 --member serviceAccount:$cloudservices_service_account --role roles/editor
gcloud compute images export --destination-uri=gs://wolk-scripts/scripts/cloudstore/wolk-image-04222019 --image=wolk-gc-us-west --project=wolk-us-west

# Import image
gcloud compute images import wolk-image --os=centos-7 --source-file=./wolk-image-04222019 --project=wolk-1307
