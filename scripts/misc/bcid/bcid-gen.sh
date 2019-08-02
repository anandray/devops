#!/bin/bash

for i in {1..10}
do
echo "
BlockChainID $i"
node index.js 0x34Dd54C7761A85f5735eEF9BD2777639418bE591 23ae5592af28db66db211de06be1c03fb009bb06cec402c714ce6451e2c520f2 &>> blockchainids
sleep 10
done
