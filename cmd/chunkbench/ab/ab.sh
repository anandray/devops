#!/bin/sh
# Usage: ab.sh cloudstore-us-central-gc.wolk.com

# POST share
printf "\n\n\n***** POST share\n"
ab -q -n 1 -c 1 -k -p chunk "http://$1/wolk/share"

# GET share
printf "\n\n\n***** GET share\n"
ab -q -n 1 -c 1 -k "http://$1/wolk/share/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"

# POST chunk
printf "\n\n\n***** POST chunk\n"
ab -q -n 1 -c 1 -k -p chunk -T 'application/json' "http://$1/wolk/chunk"

# GET chunk
printf "\n\n\n***** GET chunk\n"
ab -q -n 1 -c 1 -k "http://$1/wolk/chunk/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"
