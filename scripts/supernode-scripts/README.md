# Supernodes

## Supernode Setup 

The supernode scripts set up a Wolk mining operation on a cloud provider (Google, Azure, Amazon, Alibaba).
The inputs to the supernode script are:
 * nodenumber 
 * region

### Google 

To set up a Wolk supernode on Google Cloud:
* TODO

### Amazon

To set up a Wolk supernode on Amazon Web Services:
* TODO

### Azure

To set up a Wolk supernode on Microsoft Azure:
* TODO

### Alibaba

To set up a Wolk supernode on Alibaba Cloud:
* TODO

## nodecheck 

After your supernode is running, you can check whether it has been setup correctly with a nodecheck script.

Usage:
```# nodecheck 
Manual: (useful for testing new supernodes, not required to be be in servers table)
   nodecheck [serverip] [cloudprovider=gc|aws|az|alibaba] [nodeType=consensus|storage] (publicip)
Automatic: (useful for verifying supernodes in production, must be in servers table)
   nodecheck [serverip|all|cx|nodenumber]
   nodecheck all -- runs checks on all servers
   nodecheck 3   -- runs checks on node 3 [consensus + storage]
   nodecheck c2  -- runs checks on c2.wolk.com (must be in servers table)
   nodecheck 52.67.124.96 -- runs checks on specific ip 52.67.124.96
   nodecheck az -- runs checks on all az nodes
Checks:
 - toml presence
 - storage credentials presence
 - ssl files presence
 - nodeType consistency [consensus/storage]
 - wolk binary presence
```

Prior to registering with the Wolk blockchain:
```
# nodecheck 52.231.158.89 azure consensus
running nodecheck [server=52.231.158.89, cloudprovider=azure] ... 
52.231.158.89 [ @cloudprovider=azure,nodeType=consensus] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no 52.231.158.89 "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
52.231.158.89 TOMLCHECK: PASS wolk.toml All expected fields present
52.231.158.89 FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
52.231.158.89 FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
52.231.158.89 FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
52.231.158.89 FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present```
```


# Running a Wolk Mining Operation

To run multiple wolk mining nodes, it is useful to have a database.  At this point, we are using mysql but expect to use Wolk itself instead by the end of 2019.

If you have a project table representing multiple supernodes populating a `servers` table then this will check all of your nodes:

```
# nodecheck all
```

Checking all the nodes of the same provider: (aws, gc, azure, alibaba):
```
# nodecheck aws
c6.wolk.com [3.9.3.91 @cloudprovider=aws,nodeType=consensus] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no c6.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
c6.wolk.com TOMLCHECK: PASS wolk.toml All expected fields present
c6.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
c6.wolk.com FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
c6.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
c6.wolk.com FILECHECK: PASS AmazonCredentials /root/.aws/credentials present
c6.wolk.com FILECHECK: PASS AmazonCredentials /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present
c7.wolk.com [54.233.187.117 @cloudprovider=aws,nodeType=consensus] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no c7.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
c7.wolk.com TOMLCHECK: PASS wolk.toml All expected fields present
c7.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
c7.wolk.com FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
c7.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
c7.wolk.com FILECHECK: PASS AmazonCredentials /root/.aws/credentials present
c7.wolk.com FILECHECK: PASS AmazonCredentials /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present
s6-66ef.wolk.com [18.130.29.77 @cloudprovider=aws,nodeType=storage] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s6-66ef.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
s6-66ef.wolk.com TOMLCHECK FAIL: MISSING ConsensusIdx in wolk.toml [see timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s6-66ef.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"]
s6-66ef.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
s6-66ef.wolk.com FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
s6-66ef.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
s6-66ef.wolk.com FILECHECK: PASS AmazonCredentials /root/.aws/credentials present
s6-66ef.wolk.com FILECHECK: PASS AmazonCredentials /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present
s7-2748.wolk.com [52.67.124.96 @cloudprovider=aws,nodeType=storage] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s7-2748.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
s7-2748.wolk.com TOMLCHECK FAIL: MISSING ConsensusIdx in wolk.toml [see timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s7-2748.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"]
s7-2748.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present```
```

Checking a specific node by DNS entry:
```
# nodecheck c5
c5.wolk.com [13.93.239.206 @cloudprovider=azure,nodeType=consensus] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no c5.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
c5.wolk.com TOMLCHECK: PASS wolk.toml All expected fields present
c5.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
c5.wolk.com FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
c5.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
c5.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present```

# nodecheck s6-66ef
s6-66ef.wolk.com [18.130.29.77 @cloudprovider=aws,nodeType=storage] wolk.toml field check: timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s6-66ef.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"
s6-66ef.wolk.com TOMLCHECK FAIL: MISSING ConsensusIdx in wolk.toml [see timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no s6-66ef.wolk.com "cat /root/go/src/github.com/wolkdb/cloudstore/wolk.toml"]
s6-66ef.wolk.com FILECHECK: PASS SSLCertFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt present
s6-66ef.wolk.com FILECHECK: PASS SSLKeyFile /etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key present
s6-66ef.wolk.com FILECHECK: PASS GenesisFile /root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json present
s6-66ef.wolk.com FILECHECK: PASS AmazonCredentials /root/.aws/credentials present
s6-66ef.wolk.com FILECHECK: PASS AmazonCredentials /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk present```
```
