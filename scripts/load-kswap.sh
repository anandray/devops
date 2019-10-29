#!/bin/bash


## run as cron, thus no $PATH, thus need to define all absolute paths
top=/usr/bin/top
grep=/bin/grep


top=$($top -bn1 -o \%CPU -u0 | $grep -m2 -E "%CPU|kswapd0")

IFS='
'
set -f

i=0

for line in $top
do
        #echo $i $line

        if ! (( i++ ))
        then
                pos=${line%%%CPU*}
                pos=${#pos}
                #echo $pos
        else
                cpu=${line:(($pos-1)):3}
                cpu=${cpu// /}
                #echo $cpu
        fi

done

[[ -n $cpu ]] && \
(( $cpu >= 90 )) \
&& echo 1 > /proc/sys/vm/drop_caches \
&& echo "$$ $0: cache dropped (kswapd0 %CPU=$cpu)" >&2 \
&& exit 1

exit 0
