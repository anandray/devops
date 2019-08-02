bold=$(tput bold)
normal=$(tput sgr0)

MD5=`ssh -q yaron-phobos-a md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | awk '{print$1}'`
if ! md5sum /root/go/src/github.com/wolkdb/plasma/build/bin/plasma | grep $MD5 &> /dev/null; then

echo "
${bold}Plasma binary on yaron-phobos-a has changed. Copying binary from yaron-phobos-a...
"
echo ${normal}

scp -C -p yaron-phobos-a:/root/go/src/github.com/wolkdb/plasma/build/bin/plasma /root/go/src/github.com/wolkdb/plasma/build/bin/plasma 2> /dev/null && chmod +x /root/go/src/github.com/wolkdb/plasma/build/bin/plasma 2> /dev/null &

sleep 3

else
echo "
${bold}MD5 of Plasma binary is unchanged on yaron-phobos-a..."
echo ${normal}
fi


if ! ssh -q www6001 md5sum /var/www/vhosts/mdotm.com/httpdocs/.start/plasma | grep $MD5 &> /dev/null; then
echo "

${bold}MD5 of Plasma binary on www6001 differs from yaron-phobos-a. Copying binary from yaron-phobos-a to www6001...
"
echo ${normal}

scp -q -C -p yaron-phobos-a:/root/go/src/github.com/wolkdb/plasma/build/bin/plasma www6001:/var/www/vhosts/mdotm.com/httpdocs/.start/plasma 2> /dev/null | grep -v 'Permanently added'

sleep 3

else
echo "
${bold}MD5 of Plasma binary on www6001 and yaron-phobos-a are identical...
"
echo ${normal}

fi
