#!/bin/bash

#for i in `seq 20`; do

# to check if latency is greater than 40k ms in httpd/apache logs
latency=$(grep `date +%s --date='2 second ago'` /var/log/httpd/mdotm.com-access_log | grep -E 'chartboostbid2_3|mopubbid|adxbid|spotxbid' | grep -v '204 192' | cut -d " " -f1 | head -n1) > /dev/null

# the packet size returned
packet_size=$(grep `date +%s --date='2 second ago'` /var/log/httpd/mdotm.com-access_log | grep -E 'chartboostbid2_3|mopubbid|adxbid|spotxbid' | grep -v '204 192' | cut -d " " -f5 | head -n1) > /dev/null

# max latency in ms as reference point
max=40000

# timestamp
timestamp=$(grep `date +%s --date='2 second ago'` /var/log/httpd/mdotm.com-access_log | grep -E 'chartboostbid2_3|mopubbid|adxbid|spotxbid' | grep -v '204 192' | cut -d " " -f2 | head -n1) > /dev/null

timestamp1=`date -d@$timestamp| awk '{print$4}'`

if [ $latency -ge $max ];
then
echo "$timestamp1, $latency > $max, packet size=$packet_size"
else
echo "$timestamp1, $latency < $max, packet size=$packet_size"
fi
#fi;
#done

sh /var/www/vhosts/mdotm.com/scripts/systems/latency.sh
