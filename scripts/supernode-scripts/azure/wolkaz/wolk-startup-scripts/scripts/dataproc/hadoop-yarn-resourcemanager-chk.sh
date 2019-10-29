#!/bin/bash
ROLE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/dataproc-role)
if echo $ROLE | grep Master > /dev/null; then
sudo  /bin/systemctl restart hadoop-yarn-resourcemanager.service > /home/anand/hadoop-yarn-resourcemanager-restart.log 2>&1
#sudo /bin/sed -i 's/\*\/1 \* \* \* \* \/bin\/ssh/\#\*\/1 * * * * \/bin\/ssh/g' /var/spool/cron/crontabs/root
fi
