#!/usr/bin/env bash
 
bold=$(tput bold)
normal=$(tput sgr0)

# indicates 100%, as reference
TOT=100
# number of elements in the progress bar
ITR=10
# number of seconds to wait for
SEC=30
 
# counters
CNT=0
CUR=0
 
# for simplicity, adapts the elements to the number of seconds if they are a few
if [ $SEC -lt $ITR ];
then
   ITR=$SEC
fi
 
# a bunch of math
let SYM=ITR
let INC=TOT/ITR
let SPN=SEC/ITR
 
# prepares body and head of a printable line
HEAD='Status: ['
BODY=''
# main loop, it loops over the number of elements in the progress bar
while [ $CNT -le $ITR ];
do
   BODY=''
 
   # defines how many elements to put into the bar
   POUNDS=0
   while [ $POUNDS -lt $CNT ];
   do
       BODY="$BODY####"
       let POUNDS=POUNDS+1
   done
 
   # defines how many blank spaces to append
   let BLANKS=SYM-CNT
   while [ $BLANKS -gt 0 ];
   do
       BODY="$BODY "
       let BLANKS=BLANKS-1
   done
 
   # defines the line tail with the progress percentage
   TAIL="] ($CUR%)"
   echo -ne "$HEAD$BODY$TAIL\r"
   let CUR=CUR+INC
   let CNT=CNT+1
 
   # spins before of the next update in case its needed
   SPIN=$SPN
   while [ $SPIN -gt 0 ];
   do
       sleep 1
       let SPIN=SPIN-1
   done
done
 
# fixes any approximation caused by the integer arithmetic
if [ $CUR -ne 100 ];
then
 TAIL="] (100%)"
 echo -ne "$HEAD$BODY$TAIL\r"
fi
 
echo ''
