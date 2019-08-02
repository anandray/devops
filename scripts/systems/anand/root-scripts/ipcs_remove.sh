#!/bin/bash

s=`ipcs -s | grep -w 600 | cut -d' ' -f2`
for i in $s
do
        echo removing sem id $i
        ipcrm -s $i
done
