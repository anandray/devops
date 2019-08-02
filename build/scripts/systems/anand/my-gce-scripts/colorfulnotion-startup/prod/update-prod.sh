#!/bin/bash

        cd /home/goracing/scripts;
	kill -9 `netstat -apn | grep :9090 | awk '{print$NF}' | cut -d "/" -f1`;
	pkill -9 racerver_master;
	"yes" | rm -rf /home/goracing/go_projects/src/goracing.colorfulnotion.com/race/racerver_master;
	pkill -9 racerver_master;
	./deployProduction.sh &
