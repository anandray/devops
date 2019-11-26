MAILTO=''
SHELL=/bin/bash
BASH_ENV=/root/.bashrc

*/1 * * * * /root/scripts/geth-start-nomine.sh &> /var/log/geth-start-nomine.log
