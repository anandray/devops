#!/bin/bash
gsutil cp gs://startup_scripts_us/scripts/dataproc/fair-scheduler.xml /etc/hadoop/conf/;

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.resourcemanager.scheduler.class \
    --value org.apache.hadoop.yarn.server.resourcemanager.scheduler.fair.FairScheduler

bdconfig set_property \
    --configuration_file /etc/hadoop/conf/yarn-site.xml \
    --name yarn.scheduler.fair.allocation.file \
    --value /etc/hadoop/conf/fair-scheduler.xml

ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
if [[ "${ROLE}" == 'Master' ]]; then
  systemctl restart hadoop-yarn-resourcemanager.service
fi
