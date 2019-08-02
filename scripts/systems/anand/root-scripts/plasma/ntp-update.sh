service ntpd stop
'yes ' | cp -rf /etc/ntp.conf /etc/ntp.conf-bak
scp -q -C -p www6002:/root/scripts/plasma/ntp.conf /etc/
ntpdate pool.ntp.org
service ntpd start
