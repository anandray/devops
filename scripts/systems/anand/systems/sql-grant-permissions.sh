#!/bin/bash
gcloud compute instances list | grep -E 'www|ha6|ha2' | grep -v -E 'www600|www6010|wolk|vCPU' | awk '{print"mysql --login-path=/root/.mysql -hdb03 -e \"grant all on *.* to \"db\"\@\""$5"\" identified by \"1wasb0rn2\"\""}' | grep -v '\"db\"\@\"10\.' > /tmp/sql1.sh && sed -i "s/\"1wasb0rn2\"/\'1wasb0rn2\'/g" /tmp/sql1.sh && sh /tmp/sql1.sh

gcloud compute instances list | grep -E 'www|ha6|ha2' | grep -v -E 'www600|www6010|wolk|vCPU' | awk '{print"mysql --login-path=/root/.mysql -hdb03 -e \"grant all on *.* to \"db\"\@\""$6"\" identified by \"1wasb0rn2\"\""}' | grep -v -E 'RUNNING|TERMINATED|\"db\"\@\"10\.' > /tmp/sql2.sh && sed -i "s/\"1wasb0rn2\"/\'1wasb0rn2\'/g" /tmp/sql2.sh && sh /tmp/sql2.sh
