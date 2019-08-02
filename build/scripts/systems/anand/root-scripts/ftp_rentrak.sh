#!/bin/bash
DATE="`date +%Y-%m`"

cd /disk1/rentrak/
#ftp -v -n << "END"
ftp -v -n << EOF
open ftp.linear.rentrak.com
user mdotm_linear ToK729HaC
bin
hash
prompt
cd OUT
mget *${DATE}*
bye
EOF
