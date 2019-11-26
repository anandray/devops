sed -i 's/\*/1 \* \* \* \* \/usr\/bin\/flock -w 0 \/var\/run\/ccdexchk.lock/#\*/1 \* \* \* \* \/usr\/bin\/flock -w 0 \/var\/run\/ccdexchk.lock/g' /var/spool/cron/root
