#!/bin/bash

sudo mkdir -p /etc/pki/tls/certs/wildcard
sudo wget -O /etc/pki/tls/certs/wildcard/www.wolk.com.crt http://d5.wolk.com/.start/certs/www.wolk.com.crt &>/dev/null 2>&1 &
sudo wget -O /etc/pki/tls/certs/wildcard/www.wolk.com.key http://d5.wolk.com/.start/certs/www.wolk.com.key &>/dev/null 2>&1 &
sudo wget -O /etc/pki/tls/certs/wildcard/www.wolk.com.pem http://d5.wolk.com/.start/certs/www.wolk.com.pem &>/dev/null 2>&1 &
