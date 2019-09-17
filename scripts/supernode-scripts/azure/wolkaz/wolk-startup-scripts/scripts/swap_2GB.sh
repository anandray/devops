#!/bin/bash
sudo su - << EOF
dd if=/dev/zero of=/usr/local/swapfile_2GB bs=1024 count=2042576;
chmod 600 /usr/local/swapfile_2GB;
/sbin/mkswap -c -v1 /usr/local/swapfile_2GB;
/sbin/swapon /usr/local/swapfile_2GB;
echo "Adding entry in /etc/fstab...."
echo "/usr/local/swapfile_2GB   swap                    swap    defaults        00" >> /etc/fstab
EOF
