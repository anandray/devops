#!bin/bash
sftp -v -oIdentityFile=/root/scripts/.sftp-pass t3st@104.197.43.125 <<EOF
ls
EOF
