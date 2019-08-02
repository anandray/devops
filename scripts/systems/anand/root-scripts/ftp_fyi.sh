#!/bin/bash
DATE="`date +%Y-%m`"

cd /disk1/fyi/
#ftp -v -n << "END"
ftp -v -n << EOF
open ftp.fyitelevision.com
user mdotm EY4UuYZT
bin
hash
prompt
mget *
bye
EOF
