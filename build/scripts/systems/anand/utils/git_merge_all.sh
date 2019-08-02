#!/bin/bash
if [ -d /var/www/vhosts/$1 ]; then
cd /var/www/vhosts/$1 && git fetch upstream && git merge upstream/master
else
cd /var/www/vhosts/;
git clone git@github.com:wolktoken/$1.git;
cd $1;
git remote add upstream git@github.com:wolktoken/$1.git;
git config user.email sourabh@crosschannel.com;
git config user.name "Sourabh Niyogi";
git config --global core.filemode false;
git config core.filemode false;
git fetch upstream && git merge upstream/master;
fi
cd -
