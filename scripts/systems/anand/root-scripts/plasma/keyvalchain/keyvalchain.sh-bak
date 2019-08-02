#!/bin/bash

if ps aux | grep "keyvalchain --datadir" | grep -v grep &> /dev/null; then
kill -9 $(ps aux | grep "keyvalchain --datadir" | grep -v grep | awk '{print$2}');
sed -i 's/nohup/#nohup/g' /root/scripts/keyval-start.sh;
fi

MD5=`ssh -q 35.188.32.128 md5sum /var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain | awk '{print$1}'`

if ! md5sum /root/keyvalchain/bin/keyvalchain | grep $MD5 &> /dev/null; then
scp -C -p 35.188.32.128:/var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain /root/keyvalchain/bin/ &> /dev/null
chmod +x /root/keyvalchain/bin/keyvalchain;
#sed -i 's/#nohup/nohup/g' /root/scripts/keyval-start.sh;
fi
