#!/bin/sh
# Usage: curl.sh c0.wolk.com

# POST share
printf "\n\n\n***** POST share\n"
curl "http://$1/wolk/share" --data-binary @chunk

# getshare HTTP GET
printf "\n\n\n***** GET share\n"
curl "http://$1/wolk/share/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"

# setchunk HTTP POST
printf "\n\n\n***** POST chunk\n"
curl "http://$1/wolk/chunk/" --data-binary @chunk

# getchunk HTTP GET
printf "\n\n\n***** GET chunk\n"
curl "http://$1/wolk/chunk/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"

# setbatch HTTP POST
printf "\n\n\n***** POST batch\n"
curl "http://$1/setbatch" -d @setbatch.json

# getbatch HTTP POST
printf "\n\n\n***** POST batch\n"
curl "http://$1/getbatch" -d @getbatch.json
