#!/bin/sh

expect << 'EOS'
spawn sftp t3st@104.197.43.125:/
expect "Password:"
send "P4ssqwe1!\n"
expect "sftp>"
send "ls -l\n"
expect "sftp>"
send "bye\n"
EOS
