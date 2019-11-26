#!/bin/bash

if [ ! -d /var/www/vhosts/crosschannel.com ]; then
        echo "/var/www/vhosts/crosschannel.com does NOT exist, proceeding with git clone..."
        sudo git clone git@github.com:sourabhniyogi/crosschannel.com.git /var/www/vhosts/crosschannel.com;
        cd /var/www/vhosts/crosschannel.com/;
        git remote add upstream git@github.com:sourabhniyogi/crosschannel.com.git;
        git config user.email "sourabh@crosschannel.com";
fi

#Installing golang
if [ ! -d /usr/local/go ]; then
	sudo gsutil cp gs://startup_scripts_us/scripts/go/go1.7.1.linux-amd64.tar.gz /usr/local;
	sudo tar -C /usr/local -xzf /usr/local/go1.7.1.linux-amd64.tar.gz;
fi

#Adding environment variables to /root/.bashrc
if ! sudo grep GOPATH /root/.bashrc; then
sudo su - << EOF
sed -i '/go/d' /root/.bashrc
echo '
export PATH=$PATH:/usr/local/go/bin:/var/www/vhosts/crosschannel.com/go/bin/
export GOPATH=/var/www/vhosts/crosschannel.com/go:/var/www/vhosts/crosschannel.com/bidder
export GOROOT=/usr/local/go' >> /root/.bashrc
EOF
fi

if netstat -apn | egrep '::80' > /dev/null; then
  echo crosschannel is running
else
  echo crosschannel is NOT running
  cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php crosschannel 2> /var/log/git-crosschannel.err  > /var/log/git-crosschannel.log
fi
