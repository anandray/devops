#!/bin/bash
if ! ls -lt /var/www/vhosts/mdotm.com/httpdocs/index.php; then
rm -rf /var/www/vhosts/mdotm.com
#kill -9 `ps aux | grep git | grep -v git_clone_mdotm-new.sh | awk '{print$2}'`
cd /var/www/vhosts
git clone git@github.com:sourabhniyogi/mdotm.com.git
cd /var/www/vhosts/mdotm.com
git remote add upstream git@github.com:sourabhniyogi/mdotm.com.git
git config core.filemode false
git config user.email "sourabh@crosschannel.com"
git config user.name "Sourabh Niyogi"
git fetch upstream && git merge upstream/master
gsutil cp gs://startup_scripts_us/scripts/admin_htaccess/.htaccess /var/www/vhosts/mdotm.com/httpdocs/
gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/
fi
