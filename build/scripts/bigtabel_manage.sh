#!/bin/bash
for ((i=0; i < 16; i++)); do
         cbt -instance wlk-bt --project wolk-1307 createtable stage0-token$i
         cbt -instance wlk-bt --project wolk-1307 createfamily stage0-token$i n
         cbt -instance wlk-bt --project wolk-1307 setgcpolicy stage0-token$i n maxage=30d
done
