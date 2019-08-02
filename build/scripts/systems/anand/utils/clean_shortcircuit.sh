#!/bin/bash

if ! cat /var/www/vhosts/mdotm.com/include/shortcircuit.php | grep -n '?>' | grep '3:';
  then
    cd /var/www/vhosts/mdotm.com/;
    git checkout include/shortcircuit.php;
    git fetch 
    LOCAL=$(git rev-parse @{0})
    REMOTE=$(git rev-parse @{u})
    BASE=$(git merge-base @{0} @{u})

if [ $LOCAL = $REMOTE ]; then
   echo "no update"
  else
   echo "updating"
   git fetch upstream
   git merge upstream/master
   echo "done"
fi
    /usr/bin/gsutil cp gs://startup_scripts_us/scripts/shortcircuit.php /var/www/vhosts/mdotm.com/include/shortcircuit.php;
  else
  cat /var/www/vhosts/mdotm.com/include/shortcircuit.php
fi
