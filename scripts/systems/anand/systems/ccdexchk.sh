#!/bin/bash

psaux=`ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/ccdex' | grep -v grep`
psaux_cc=`ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/crosschannel' | grep -v grep`

#if echo $psaux_cc &> /dev/null; then
if ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/crosschannel' | grep -v grep &> /dev/null; then
echo 
"-------------------------------------
crosschannel is running, killing...
-------------------------------------
"

#kill -9 $(echo $psaux_cc | awk '{print$2}')
kill -9 $(ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/crosschannel' | grep -v grep | awk '{print$2}')
fi

# checking curl call
if ! curl -s "http://127.0.0.1/data/healthcheck" | grep OK > /dev/null; then
echo "
`date +%m-%d-%T`:
-------------------------------------
ccdex is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_ccdex.sh
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
ccdex is running...

$psaux
---------------------------------------------------------------------------------------------------------------
"
fi

if ! ps aux | grep '/var/www/vhosts/crosschannel.com/bidder/bin/ccdex' | grep -v grep &> /dev/null; then
echo "
`date +%m-%d-%T`:
-------------------------------------
ccdex is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_ccdex.sh
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
ccdex is running... 

$psaux
---------------------------------------------------------------------------------------------------------------
"
fi


if sh /var/www/vhosts/mdotm.com/scripts/systems/port_80.sh | grep 'Escape character is' &> /dev/null; then
echo "
`date +%m-%d-%T`:
----------------------------------
ccdex is running on port 80
----------------------------------
"
else
echo "
`date +%m-%d-%T`:
-------------------------------------
ccdex is NOT running.. Restarting....
-------------------------------------
"
sh /var/www/vhosts/mdotm.com/scripts/systems/run_ccdex.sh
echo "
`date +%m-%d-%T`:
---------------------------------------------------------------------------------------------------------------
$psaux
---------------------------------------------------------------------------------------------------------------
"
fi
