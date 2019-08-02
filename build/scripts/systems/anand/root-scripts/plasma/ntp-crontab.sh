echo "
*/5 * * * * service ntpd stop;ntpdate -v -u -b pool.ntp.org;service ntpd start" >> /var/spool/cron/root
