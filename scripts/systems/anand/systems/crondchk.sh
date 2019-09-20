#!/bin/sh
LIMIT1=100
LIMIT2=25
Load_AVG=`uptime | cut -d'l' -f2 | awk '{print $3}' | cut -d. -f1`

for i in {1..12};
do

if [ $Load_AVG -lt $LIMIT2 ] && ! service crond status | grep running &> /dev/null; then
echo "`date +%m-%d-%Y-%T`"
/sbin/service crond restart;
else
echo "`date +%m-%d-%Y-%T` - crond is running..."
fi
sleep 5;
done