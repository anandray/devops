#!/bin/bash
YEAR=`date +%Y`
#find /var/log -name "*$YEAR*" -exec du -sh {} \;
find /var/log -name "*$YEAR*" -exec rm -rfv {} \;
