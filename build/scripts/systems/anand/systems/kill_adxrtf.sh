#!/bin/bash

/bin/kill -9 `ps aux | grep adxrtf.php | grep -v grep | awk '{print$2}'`
