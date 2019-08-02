#!/bin/bash

#------------ GLIBC  UPDATE ------------
#updated GLIBC to run downloaded GETH BINARY:
#http://stackoverflow.com/questions/35616650/how-to-upgrade-glibc-from-version-2-12-to-2-14-on-centos

if [ ! -d /opt/glibc-2.14 ]; then
mkdir ~/glibc_install; cd ~/glibc_install
wget http://ftp.gnu.org/gnu/glibc/glibc-2.14.tar.gz;
tar zxvf glibc-2.14.tar.gz;
cd ~/glibc_install/glibc-2.14;
mkdir ~/glibc_install/glibc-2.14/build;
cd ~/glibc_install/glibc-2.14/build
../configure --prefix=/opt/glibc-2.14
make -j4
sudo make install
export LD_LIBRARY_PATH=/opt/glibc-2.14/lib

echo 'export LD_LIBRARY_PATH=/opt/glibc-2.14/lib' >> ~/.bash_profile
source ~/.bash_profile
else
echo "/opt/glibc-2.14 is already installed..."
fi
#------------ /END GLIBC  UPDATE ------------

if ! which geth &> /dev/null; then
# Download geth and other tools binaries
mkdir ~/ethereum
cd ~/ethereum
wget -c https://gethstore.blob.core.windows.net/builds/geth-alltools-linux-amd64-1.6.0-facc47cb.tar.gz;
tar zxvpf geth-alltools-linux-amd64-1.6.0-facc47cb.tar.gz;
cp -rf ~/ethereum/geth-alltools-linux-amd64-1.6.0-facc47cb/geth /usr/bin/;
chmod +x /usr/bin/geth
else
echo "geth is already installed.."
fi

if ! which solc &> /dev/null; then
# Download solc binary
cd ~/ethereum
wget -c https://github.com/ethereum/solidity/releases/download/v0.4.10/solc;
cp ~/ethereum/solc /usr/bin
chmod +x /usr/bin/solc
else
echo "solc is already installed..."
fi
