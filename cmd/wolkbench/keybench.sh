# keybench
nohup wolkbench -server=c0.wolk.com -httpport=443 -users=2 -files=2 -run=key &> /var/log/wolkbench.log &
nohup wolkbench -server=c0.wolk.com -httpport=81  -users=2 -files=3 -run=key &> /var/log/wolkbench1.log &
nohup wolkbench -server=c0.wolk.com -httpport=82  -users=1 -files=1 -run=key &> /var/log/wolkbench2.log &
nohup wolkbench -server=c0.wolk.com -httpport=83  -users=4 -files=4 -run=key &> /var/log/wolkbench3.log &
nohup wolkbench -server=c1.wolk.com -httpport=84  -users=3 -files=3 -run=key &> /var/log/wolkbench4.log &
nohup wolkbench -server=c1.wolk.com -httpport=85  -users=4 -files=4 -run=key &> /var/log/wolkbench5.log &

# filebench
# nohup wolkbench -server=c0.wolk.com         -httpport=81  -users=3 -files=3   -filesize=10000   -run=file &> /var/log/wolkbench1.log
# nohup wolkbench -server=c0.wolk.com         -httpport=82  -users=10 -files=10 -filesize=100000  -run=file &> /var/log/wolkbench2.log
# nohup wolkbench -server=cloudflare.wolk.com -httpport=83  -users=20 -files=20 -filesize=1000000 -run=file &> /var/log/wolkbench3.log
