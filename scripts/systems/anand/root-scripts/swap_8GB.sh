#!/bin/bash
dd if=/dev/zero of=/usr/local/swapfile_8GB bs=1024 count=8194304;
chmod 600 /usr/local/swapfile_8GB;
/sbin/mkswap -c -v1 /usr/local/swapfile_8GB;
/sbin/swapon /usr/local/swapfile_8GB;
echo "Adding entry in /etc/fstab...."
echo "/usr/local/swapfile_8GB   swap                    swap    defaults        00" >> /etc/fstab
