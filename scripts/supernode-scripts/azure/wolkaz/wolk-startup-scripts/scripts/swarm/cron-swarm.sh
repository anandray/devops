MAILTO=''
SHELL=/bin/bash
BASH_ENV=/root/.bashrc

*/1 * * * * /bin/sh /root/scripts/geth-install.sh &> /var/log/geth.log
*/1 * * * * sh /root/scripts/geth-start.sh &>> /var/log/geth-start.log