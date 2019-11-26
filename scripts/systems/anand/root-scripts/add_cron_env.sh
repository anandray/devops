#!/bin/bash
sed -i '3 i\BASH_ENV=\/etc\/profile.d\/cron_env.bash' /var/spool/cron/root
