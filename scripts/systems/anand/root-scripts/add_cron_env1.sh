#!/bin/bash
sed -i 's/BASH_ENV=\/root\/.bashrc/\#BASH_ENV=\/root\/.bashrc/g' /var/spool/cron/root
