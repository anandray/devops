# c cluster (stage 2)
/root/go/src/github.com/wolkdb/wolkjs/wcloud createaccount --httpport=82 --server=c0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud getname       --httpport=82 --server=c0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=82 --server=c0.wolk.com wolk://user123/planets
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=82 --server=c0.wolk.com wolk://user123/planets/pluto  dwarfy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=82 --server=c0.wolk.com wolk://user123/planets/saturn ringy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=82 --server=c0.wolk.com wolk://user123/planets/earth  watery
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=82 --server=c0.wolk.com wolk://user123/planets/saturn
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=82 --server=c0.wolk.com wolk://user123/planets/
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=82 --server=c0.wolk.com wolk://user123/videos
/root/go/src/github.com/wolkdb/wolkjs/wcloud put           --httpport=82 --server=c0.wolk.com /root/go/src/github.com/wolkdb/devops/testnet/gilroy.mp4  wolk://user123/videos/gilroy.mp4
# keybench with 2*2 
# filebench with 1*1 1MB 
# ... fuse kubecluster branch into dev

# z cluster (gc)
/root/go/src/github.com/wolkdb/wolkjs/wcloud createaccount --httpport=2080 --server=z0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud getname       --httpport=2080 --server=z0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=2080 --server=z0.wolk.com wolk://user123/planets
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=2080 --server=z0.wolk.com wolk://user123/planets/pluto  dwarfy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=2080 --server=z0.wolk.com wolk://user123/planets/saturn ringy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=2080 --server=z0.wolk.com wolk://user123/planets/earth  watery
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=2080 --server=z0.wolk.com wolk://user123/planets/saturn
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=2080 --server=z0.wolk.com wolk://user123/planets/
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=2080 --server=z0.wolk.com wolk://user123/videos
/root/go/src/github.com/wolkdb/wolkjs/wcloud put           --httpport=2080 --server=z0.wolk.com /root/go/src/github.com/wolkdb/devops/testnet/gilroy.mp4  wolk://user123/videos/gilroy.mp4
# wolkbench with 2*2
# filebench with 1*1 1MB 

# w cluster (aws)
/root/go/src/github.com/wolkdb/wolkjs/wcloud createaccount --httpport=3080 --server=w0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud getname       --httpport=3080 --server=w0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=3080 --server=w0.wolk.com wolk://user123/planets
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=3080 --server=w0.wolk.com wolk://user123/planets/pluto  dwarfy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=3080 --server=w0.wolk.com wolk://user123/planets/saturn ringy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=3080 --server=w0.wolk.com wolk://user123/planets/earth  watery
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=3080 --server=w0.wolk.com wolk://user123/planets/saturn
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=3080 --server=w0.wolk.com wolk://user123/planets/
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=3080 --server=w0.wolk.com wolk://user123/videos
/root/go/src/github.com/wolkdb/wolkjs/wcloud put           --httpport=3080 --server=w0.wolk.com /root/go/src/github.com/wolkdb/devops/testnet/gilroy.mp4  wolk://user123/videos/gilroy.mp4
# wolkbench with 2*2
# filebench with 1*1 1MB 

# v cluster (azure)
/root/go/src/github.com/wolkdb/wolkjs/wcloud createaccount --httpport=4080 --server=v0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud getname       --httpport=4080 --server=v0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=4080 --server=v0.wolk.com wolk://user123/planets
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=4080 --server=v0.wolk.com wolk://user123/planets/pluto  dwarfy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=4080 --server=v0.wolk.com wolk://user123/planets/saturn ringy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=4080 --server=v0.wolk.com wolk://user123/planets/earth  watery
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=4080 --server=v0.wolk.com wolk://user123/planets/saturn
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=4080 --server=v0.wolk.com wolk://user123/planets/
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=4080 --server=v0.wolk.com wolk://user123/videos
/root/go/src/github.com/wolkdb/wolkjs/wcloud put           --httpport=4080 --server=v0.wolk.com /root/go/src/github.com/wolkdb/devops/testnet/gilroy.mp4  wolk://user123/videos/gilroy.mp4
# wolkbench with 2*2
# filebench with 1*1 1MB 

# x cluster (alibaba)
/root/go/src/github.com/wolkdb/wolkjs/wcloud createaccount --httpport=1080 --server=x0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud getname       --httpport=1080 --server=x0.wolk.com user123
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=1080 --server=x0.wolk.com wolk://user123/planets
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=1080 --server=x0.wolk.com wolk://user123/planets/pluto  dwarfy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=1080 --server=x0.wolk.com wolk://user123/planets/saturn ringy
/root/go/src/github.com/wolkdb/wolkjs/wcloud set           --httpport=1080 --server=x0.wolk.com wolk://user123/planets/earth  watery
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=1080 --server=x0.wolk.com wolk://user123/planets/saturn
/root/go/src/github.com/wolkdb/wolkjs/wcloud get           --httpport=1080 --server=x0.wolk.com wolk://user123/planets/
/root/go/src/github.com/wolkdb/wolkjs/wcloud mkdir         --httpport=1080 --server=x0.wolk.com wolk://user123/videos
/root/go/src/github.com/wolkdb/wolkjs/wcloud put           --httpport=1080 --server=x0.wolk.com /root/go/src/github.com/wolkdb/devops/testnet/gilroy.mp4  wolk://user123/videos/gilroy.mp4
# wolkbench with 2*2
# filebench with 1*1 1MB 






