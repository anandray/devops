MAILTO=''
SHELL=/bin/bash
BASH_ENV=/root/.bashrc

*/1 * * * * /root/scripts/geth-ropsten-start-nomine.sh &> /var/log/geth-ropsten-start-nomine.log
