#!/bin/bash

#SSH Keys:
if [ ! -f /root/.ssh/authorized_keys ]; then
        sudo gsutil cp gs://startup_scripts_us/scripts/ssh_keys.tgz /root/.ssh/
        sudo tar zxvpf /root/.ssh/ssh_keys.tgz -C /root/.ssh/
        sudo rm -rf /root/.ssh/ssh_keys.tgz
else

	echo "/root/.ssh/authorized_keys exists..."
	sed -i 's/\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/ssh_keys_chk.sh/\#\*\/1 \* \* \* \* \/bin\/sh \/root\/scripts\/ssh_keys_chk.sh/g' /var/spool/cron/root
fi
