#!/bin/bash

psaux=`ps aux | grep '/var/www/vhosts/api.wolk.com/bin/wolk' | grep -v grep`

# checking curl call
if ! curl -s "http://127.0.0.1/data/healthcheck" | grep OK > /dev/null; then
echo "
`date +%m-%d-%T`:
-------------------------------------
wolk is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_wolk.sh
echo "
`date +%m-%d-%T`:
-------------------------------------
$psaux
-------------------------------------
"
else
echo "
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------------------
wolk is running...

$psaux
---------------------------------------------------------------------------------------------------------------
"
fi

if ! ps aux | grep '/var/www/vhosts/api.wolk.com/bin/wolk' | grep -v grep &> /dev/null; then
echo "
`date +%m-%d-%T`:
-------------------------------------
wolk is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_wolk.sh
echo "
`date +%m-%d-%T`:
-------------------------------------
$psaux
-------------------------------------
"
else
echo "
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------------------
wolk is running... 

$psaux
---------------------------------------------------------------------------------------------------------------
"
fi


if sh /var/www/vhosts/mdotm.com/scripts/systems/port_80.sh | grep 'Escape character is' &> /dev/null; then
echo "
`date +%m-%d-%T`:
----------------------------------
wolk is running on port 80
----------------------------------
"
else
echo "
`date +%m-%d-%T`:
-------------------------------------
wolk is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_wolk.sh
echo "
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------------------
$psaux
---------------------------------------------------------------------------------------------------------------
"
fi
