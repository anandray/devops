#!/bin/bash
rm -rf /var/www/vhosts/api.wolk.com  &&
cd /var/www/vhosts  &&
git clone git@github.com:wolktoken/api.wolk.com.git  &&
cd api.wolk.com  &&
git remote add upstream git@github.com:wolktoken/api.wolk.com.git  &&
git config user.email sourabh@wolk.com  &&
git config user.name "Sourabh Niyogi"  &&
git config --global core.filemode false  &&
git config core.filemode false  &&
git fetch upstream && git merge upstream/master
