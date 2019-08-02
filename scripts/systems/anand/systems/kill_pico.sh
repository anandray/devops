#!/bin/bash

/bin/kill -9 `ps aux | grep pico.php | grep -v grep | awk '{print$2}'`
