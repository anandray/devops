#!/bin/bash
mkdir -v -p /disk1/home/sftp/$1/incoming
echo "
"
useradd -G sftpusers -g sftpusers -d /disk1/home/sftp/$1 -s /sbin/nologin $1 2> /dev/null

chown -v $1.sftpusers /disk1/home/sftp/$1/incoming
echo "
"
chown -v root.root /disk1/home/sftp/$1
echo "
"
chmod -v 0755 /disk1/home/sftp/$1
passwd $1

echo "
Ownership should be root.root for /disk1/home/sftp/$1
"
ls -ld /disk1/home/sftp/$1/

echo "
Ownership should be $1.sftpusers for /disk1/home/sftp/$1/incoming
"
ls -ld /disk1/home/sftp/$1/incoming

echo "
Making sure there are no .bash files in home_dir
"
ls -la /disk1/home/sftp/$1/

echo "
Username: $1
Password: _______________
SFTP Address/Host: $1.crosschannel.com
"
