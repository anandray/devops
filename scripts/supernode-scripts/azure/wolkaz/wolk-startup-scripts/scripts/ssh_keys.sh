#!/bin/bash
#sudo wget -P /root/.ssh/ http://anand.www1001.mdotm.com/gce/ssh_keys/ssh_keys.tgz
gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz .ssh/ && 
sudo cp /home/anand/.ssh/ssh_keys.tgz /root/.ssh/ && 
sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
