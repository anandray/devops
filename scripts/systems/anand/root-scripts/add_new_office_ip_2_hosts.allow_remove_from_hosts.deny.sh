#!/bin/bash

echo "# new office ip - 55 E 3rd Ave
50.202.37.133
" >> /var/lib/denyhosts/allowed-hosts
service denyhosts restart
