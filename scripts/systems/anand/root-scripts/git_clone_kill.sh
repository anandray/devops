#!/bin/bash
kill -9 `ps aux | grep git | awk '{print$2}'`
