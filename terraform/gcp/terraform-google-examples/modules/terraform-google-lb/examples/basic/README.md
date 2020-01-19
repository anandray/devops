# TCP Forwarding Rule Example

[![button](http://gstatic.com/cloudssh/images/open-btn.png)](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/GoogleCloudPlatform/terraform-google-lb&working_dir=examples/basic&page=shell&tutorial=README.md)

<a href="https://concourse-tf.gcp.solutions/teams/main/pipelines/tf-examples-lb-basic" target="_blank">
<img src="https://concourse-tf.gcp.solutions/api/v1/teams/main/pipelines/tf-examples-lb-basic/badge" /></a>

This example creates a managed instance group with 2 instances in the same region and a network TCP Load Balancer.

**Figure 1.** *diagram of Google Cloud resources*

![architecture diagram](https://raw.githubusercontent.com/GoogleCloudPlatform/terraform-google-lb/master/examples/basic/diagram.png)

## Change to the example directory

```
[[ `basename $PWD` != basic ]] && cd examples/basic
```

## Install Terraform

1. Install Terraform if it is not already installed (visit [terraform.io](https://terraform.io) for other distributions):

```
../terraform-install.sh
```

## Set up the environment

1. Set the project, replace `YOUR_PROJECT` with your project ID:

```
PROJECT=YOUR_PROJECT
```

```
gcloud config set project ${PROJECT}
```

2. Configure the environment for Terraform:

```
[[ $CLOUD_SHELL ]] || gcloud auth application-default login
export GOOGLE_PROJECT=$(gcloud config get-value project)
```

## Run Terraform

```
terraform init
terraform apply
```

## Testing

1. Wait for the load balancer to be provisioned:

```
./test.sh
```

2. Open the URL of the load balancer in your browser:

```
echo http://$(terraform output load-balancer-ip)
```

You should see the instance details from `group1`

## Cleanup

1. Remove all resources created by terraform:

```
terraform destroy
```
