#!/bin/bash

for i in {1..12};
do
if ps aux | grep 'node /usr/local/slack/node_modules/slackin/bin/slackin wolktoken xoxp-189648714321-213522326502-228047900416-859aa0a30a9f7055c333b8f9e34dd6e1 6Ld8CikUAAAAAIGGQH3RascRCURamSxcCIoHauIe 6Ld8CikUAAAAADc7WiEGef2u-IgHm0zKd1QCqTuq -p 1450 --css /assets/slack.css' | grep -v grep &> /dev/null; then
echo "`date +'%m%d%Y %T'` - Slackin is running..."
else
echo "`date +'%m%d%Y %T'` - Slackin is NOT running... restarting"
/usr/local/slack/node_modules/slackin/bin/slackin wolktoken xoxp-189648714321-213522326502-228047900416-859aa0a30a9f7055c333b8f9e34dd6e1 6Ld8CikUAAAAAIGGQH3RascRCURamSxcCIoHauIe 6Ld8CikUAAAAADc7WiEGef2u-IgHm0zKd1QCqTuq -p 1450 --css /assets/slack.css & /sbin/service httpd restart;
fi
sleep 5;
done
