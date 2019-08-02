#!/bin/bash
if ! docker container ls -a | grep wolk &> /dev/null; then
echo "`date +%T` - No Docker container found.. starting..."
docker run --name=wolkmain --rm -it -p 8500:8500 -p 5001:5000 wolkmain &
fi
