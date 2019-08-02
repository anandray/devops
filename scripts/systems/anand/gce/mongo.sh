#!/bin/bash

# installing openssl libraries required for mongo php extensions
sudo yum -y install openssl openssl-devel;

# mongo php extensions
sudo pecl channel-update pecl.php.net;
sudo pecl install mongo;
sudo pecl install mongodb;

# removing extensions from php.ini before adding to avoud duplicacies
sudo sed -i '/mongo/d' /etc/php.ini;
sudo sed -i '/ffmpeg/d' /etc/php.ini;
sudo sed -i '/igbinary/d' /etc/php.ini;

# adding php extensions to php.ini
sudo su - << EOF
echo 'extension=mongo.so' >> /etc/php.ini;
echo 'extension=mongodb.so' >> /etc/php.ini;
echo '
extension=igbinary.so
session.serialize_handler=igbinary
igbinary.compact_strings=On' >> /etc/php.ini
EOF
sudo service httpd restart;
