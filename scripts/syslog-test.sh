#!/bin/bash

cd /root/go/src/github.com/wolkdb/cloudstore/log; git checkout log-streams; go test -run TestLogStages; git checkout dev
