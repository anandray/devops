#!/bin/bash
if ! ps aux | grep google_accounts_daemon | grep -v grep &> /dev/null; then
echo "`date +'%m%d%Y %T'` - google_accounts_daemon NOT running.. Restarting"
/usr/bin/python /usr/bin/google_accounts_daemon &
else
echo "`date +'%m%d%Y %T'` - google_accounts_daemon running"
fi

if ! ps aux | grep google_clock_skew_daemon | grep -v grep &> /dev/null; then
echo "`date +'%m%d%Y %T'` - google_clock_skew_daemon NOT running.. Restarting"
/usr/bin/python /usr/bin/google_clock_skew_daemon &
else
echo "`date +'%m%d%Y %T'` - google_clock_skew_daemon running"
fi

if ! ps aux | grep google_ip_forwarding_daemon | grep -v grep &> /dev/null; then
echo "`date +'%m%d%Y %T'` - google_ip_forwarding_daemon NOT running.. Restarting"
/usr/bin/python /usr/bin/google_ip_forwarding_daemon &
else
echo "`date +'%m%d%Y %T'` - google_ip_forwarding_daemon running"
fi
