#!/bin/bash
cd /var/www/vhosts/crosschannel.com
git fetch upstream && git merge upstream/master
#cd /var/www/vhosts/crosschannel.com/bidder/bin && php goservice.php bt
cd /var/www/vhosts/crosschannel.com/bidder/bin && sh goservice.sh bt
