#!/bin/bash
#sed -i 's/BASH_ENV=\/root\/.bashrc/\#BASH_ENV=\/root\/.bashrc/g' /var/spool/cron/root
sed -i 's/BASH_ENV=\/etc\/profile.d\/cron_env.bash/BASH_ENV=\/root\/scripts\/cron_env.bash/g' /var/spool/cron/root
