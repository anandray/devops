#!/bin/bash

hostname="ftp.linear.rentrak.com"
username="mdotm_linear"
password="ToK729HaC"

function LsFiles(){
 ftp -v -n << "end_ftp"
	open ftp.linear.rentrak.com
 	user mdotm_linear ToK729HaC
  	cd OUT
	ls -la 
	bye
end_ftp
}

LsFiles | grep "$@"  | awk '{print $NF}' > /tmp/rentrak.txt

function getFiles(){
 ftp -i -v -n << "end_ftp"
        open ftp.linear.rentrak.com
	user mdotm_linear ToK729HaC
	cd OUT
	get $1 /disk1/rentrak/
	bye
end_ftp
}

for run in $(cat /tmp/rentrak.txt)
do
getFiles $run
done
