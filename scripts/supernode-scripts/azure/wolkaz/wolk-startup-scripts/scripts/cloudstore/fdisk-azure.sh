#/bin/bash

sudo fdisk /dev/sda << EOF
d
2
n
p
2


w
EOF

sudo reboot
