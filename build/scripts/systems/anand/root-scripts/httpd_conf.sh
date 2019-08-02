scp 10.128.0.56:/etc/httpd/conf/httpd1.conf /etc/httpd/conf/httpd.conf
echo '
LogFormat "%D %{%s}t \"%U\" %>s %O \"%{HTTP_X_FORWARDED_FOR}e\"" smcommon' >> /etc/httpd/conf/httpd.conf &&

echo "

NameVirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80

<VirtualHost `ifconfig eth0| grep 'inet addr:' | cut -d: -f2| awk '{print$1}'`:80>
  ServerName `hostname`
  ServerAlias www.mdotm.com rtb-adx-eu.mdotm.com rtb-adx-eu.mdotm.co ads.mdotm.com secure.mdotm.com
  DocumentRoot /var/www/vhosts/mdotm.com/httpdocs
  ErrorLog logs/mdotm.com-error_log
  CustomLog logs/mdotm.com-access_log smcommon
</VirtualHost>

SetEnv mach `hostname -s`
SetEnv eu true
SetEnv rtb true
SetEnv adx true
SetEnv sj true
SetEnv rtb true
SetEnv wdc true
SetEnv wdc2 true

<Directory />
 Options All
    AllowOverride All
</Directory>

ExtendedStatus On

<Location /server-status>
    SetHandler server-status
    Order Deny,Allow
    Deny from all
    Allow from 127.0.0.1 10.84.81.165 75.126.67.187
</Location>
ServerName `hostname`:80" >> /etc/httpd/conf/httpd.conf;
service httpd restart;
