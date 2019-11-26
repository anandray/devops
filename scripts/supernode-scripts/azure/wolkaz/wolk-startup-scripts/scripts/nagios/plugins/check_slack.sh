#!/bin/bash
STATE_OK=0
STATE_WARNING=1
STATE_CRITICAL=2
STATE_UNKNOWN=3
 
SLACK_CHECK=`curl -s "https://slack.wolk.com/" | grep 'Join Wolk Token on Slack!' | wc -l`
 
case "${SLACK_CHECK}" in
        0)  echo "Slack is not running - run \"pkill -9 node && sh /root/scripts/systems/slackin_chk.sh\""; exit ${STATE_CRITICAL}
        ;;
        1)  echo "Slack is running - OK"; exit ${STATE_OK}
        ;;
        *)  echo "Slack is in an unknown state."; exit ${STATE_WARNING}
        ;;
esac
