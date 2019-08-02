#!/bin/bash
if ! date | grep -E 'PST|PDT';
  then
    ln -sf /usr/share/zoneinfo/America/Los_Angeles /etc/localtime;
#    service ntpd stop;
    ntpdate -u pool.ntp.org;
#    service ntpd start;

  else
#    service ntpd stop;
    ntpdate -u pool.ntp.org;
#    service ntpd start;
fi
