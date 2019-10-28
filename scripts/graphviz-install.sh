cd /root
wget https://graphviz.gitlab.io/pub/graphviz/stable/SOURCES/graphviz.tar.gz
tar zxvpf graphviz.tar.gz
rm -rf graphviz.tar.gz
mv -f graphviz* /opt/
cd /opt

# install Criterion for dependency
yum -y install cmake libgit2-devel gcc-c++
git clone --recursive https://github.com/Snaipe/Criterion
cd Criterion
mkdir build
cd build
cmake ..
cmake --build .
make install

cd /opt/graphviz*
aclocal
automake
./configure
make && make install
