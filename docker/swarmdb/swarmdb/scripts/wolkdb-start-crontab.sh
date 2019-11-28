#!/bin/bash
echo "*/1 * * * * /usr/local/swarmdb/scripts/wolkdb-start.sh &>> /usr/local/swarmdb/log/wolkdb-start.log" >> /var/spool/cron/root
