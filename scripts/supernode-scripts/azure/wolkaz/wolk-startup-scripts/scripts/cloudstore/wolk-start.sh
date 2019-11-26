#!/bin/bash

# enable wolk service
if ! ps aux | grep wolk | grep -v grep; then
echo "
WOLK is NOT running.. Starting WOLK...
"
sudo mkdir /usr/local/wolk /usr/local/cloudstore
sudo chkconfig wolk on
sudo systemctl restart wolk.service
fi
