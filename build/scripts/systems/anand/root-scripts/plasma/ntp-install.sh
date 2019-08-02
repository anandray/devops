if ! rpm -qa | grep 'ntp-' | grep -v grep &> /dev/null; then
yum -y install ntp ntpdate rdate;
fi

chkconfig ntpd on
if ! grep pool.ntp.org /etc/ntp.conf &> /dev/null; then
service ntpd stop
'yes' | cp -rf /etc/ntp.conf /etc/ntp.conf-bak
scp -q -C -p www6002:/root/scripts/plasma/ntp.conf /etc/
ntpdate -u -b pool.ntp.org
service ntpd start
else
service ntpd stop
ntpdate -u -b pool.ntp.org
service ntpd start
fi
