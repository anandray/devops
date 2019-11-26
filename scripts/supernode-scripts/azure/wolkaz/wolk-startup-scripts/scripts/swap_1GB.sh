#!/bin/bash
sudo su - << EOF
dd if=/dev/zero of=/usr/local/swapfile_1GB bs=1024 count=1021288;
chmod 600 /usr/local/swapfile_1GB;
/sbin/mkswap -c -v1 /usr/local/swapfile_1GB;
/sbin/swapon /usr/local/swapfile_1GB;
echo "Adding entry in /etc/fstab...."
echo "/usr/local/swapfile_1GB   swap                    swap    defaults        00" >> /etc/fstab
EOF
