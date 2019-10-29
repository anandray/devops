#!/bin/bash
gethaddr=`cat /var/log/geth-account-new.log | grep Address | cut -d "{" -f2 | cut -d "}" -f1`
echo '{
  "config": {
        "chainId": 1337,
        "homesteadBlock": 0,
        "eip155Block": 0,
        "eip158Block": 0
  },
  "nonce": "0x0000000000000042",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "difficulty": "0x4000",
  "alloc": {
    "0x12233992092d7b405355d771940e5115c17f959f": {
    "balance": "1000000000000000000000"
    }
  },
  "coinbase": "0x0000000000000000000000000000000000000000",
  "timestamp": "0x00",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "extraData": "0x00",
  "gasLimit": "0xffffffff"
}' > /var/www/vhosts/data/genesis.json

sudo gsutil cp /var/www/vhosts/data/genesis.json gs://startup_scripts_us/scripts/swarm/
