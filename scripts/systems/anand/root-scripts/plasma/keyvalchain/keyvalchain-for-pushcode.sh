#### start of keyvalchain.sh ####
if [ ! -d /root/keyvalchain/bin ]; then
mkdir -p /root/keyvalchain/bin
fi

if [ ! -d /root/keyvalchain/qdata/dd ]; then
mkdir -p /root/keyvalchain/qdata/dd
fi

#scp -C -p 35.188.32.128:/var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain /root/keyvalchain/bin/keyvalchain

bold=$(tput bold)
normal=$(tput sgr0)

MD5=`ssh -q 35.188.32.128 md5sum /var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain | awk '{print$1}'`

if ! md5sum /root/keyvalchain/bin/keyvalchain | grep $MD5 &> /dev/null; then

echo "
${bold}Plasma binary on dev-mayumi-go-ethereum-dqpp has changed. Copying binary from dev-mayumi-go-ethereum-dqpp...
"
echo ${normal}

scp -C -p 35.188.32.128:/var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain /root/keyvalchain/bin/keyvalchain 2> 
/dev/null && chmod +x /root/keyvalchain/bin/keyvalchain 2> /dev/null &

sleep 3

else
echo "
${bold}MD5 of Plasma binary is unchanged on dev-mayumi-go-ethereum-dqpp..."
echo ${normal}
fi


if ! ssh -q www6001 md5sum /var/www/vhosts/mdotm.com/httpdocs/.start/keyvalchain | grep $MD5 &> /dev/null; then
echo "

${bold}MD5 of Plasma binary on www6001 differs from dev-mayumi-go-ethereum-dqpp. Copying binary from dev-mayumi-go-ethereum-dqpp
 to www6001...
"
echo ${normal}

scp -q -C -p 35.188.32.128:/var/www/vhosts/src/github.com/wolkdb/plasma/build/bin/keyvalchain www6001:/var/www/vhosts/mdotm.com/
httpdocs/.start/keyvalchain 2> /dev/null | grep -v 'Permanently added'

sleep 3

else
echo "
${bold}MD5 of Plasma binary on www6001 and dev-mayumi-go-ethereum-dqpp are identical...
"
echo ${normal}

fi
#### end of keyvalchain.sh ####
