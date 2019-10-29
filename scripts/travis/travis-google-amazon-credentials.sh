#!/bin/bash

sudo wget -O ~/gopath/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json http://d5.wolk.com/.start/.a335ab6b/.a7f8f2399f8
#sudo wget -O ~/gopath/src/github.com/wolkdb/cloudstore/wolk.toml http://d5.wolk.com/.start/.a335ab6b/.wolk.toml
#sed -i '/GoogleDatastoreCredentials/d' ~/gopath/src/github.com/wolkdb/cloudstore/wolk.toml
echo 'GoogleDatastoreCredentials  = "~/gopath/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/google.json"' >> ~/gopath/src/github.com/wolkdb/cloudstore/wolk.toml;
