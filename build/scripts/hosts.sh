#!/bin/bash
echo -e "\nUsage: hosts.sh [c|s]"
mysql --defaults-extra-file=~/.mysql wolk -e "select dns from servers order by nodenumber" | awk '{print$1".wolk.com"}' | grep ^$1
