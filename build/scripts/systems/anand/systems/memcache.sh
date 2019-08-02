#!/bin/sh
if ps aux | grep mem| awk '{print$14}' | grep 83  > /dev/null; then
#if ps -A -U root | grep memcached > /dev/null; then
#if ps -A -U root | grep memcached > /root/scripts/memcached.log; then
echo memcached is running
else
echo memcached is NOT running
#kill -9 `pidof memcached`
#killall memcached
/etc/init.d/memcached stop
/etc/init.d/memcached start
/bin/ps aux | grep memcached > /root/scripts/memcached.log
kill -9 `ps auxw |grep block-map|grep -v grep| awk '{print $2}'`
php /var/www/vhosts/mdotm.com/scripts/systems/datadog-event.php memcache
cat /home/activeusers/activeusers.txt| php /var/www/vhosts/mdotm.com/hadoop/profile/block-map.php >> /var/log/recovermemcache.log 2>/dev/null &
fi
