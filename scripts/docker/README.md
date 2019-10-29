# Install Docker CE (Community Edition)
https://www.docker.com/community-edition#/download

### CentOS:
  - Installation instructions: https://docs.docker.com/engine/installation/linux/docker-ce/centos/

### Mac:
  - Installation instructions: https://store.docker.com/editions/community/docker-ce-desktop-mac

### Others:
  - https://www.docker.com/community-edition#/download
  
# System Prerequisites

|OS| Prerequisite |
|--|:--|
|CentOS|7.x (64-bit)|
|RedHat|RHEL 7.x (64-bit)|
|Debian|Stretch, Jessie 8.0, Wheezy 7.7 (64-bit)|
|Fedora|Fedora 25, Fedora 24 (64-bit)|
|Ubuntu|Zesty 17.04 (LTS),Yakkety 16.10, Xenial 16.04 (LTS),Trusty 14.04 (LTS)|
|OSX|Yosemite 10.11 or above|
|MS|Windows 10 Professional or Enterprise (64-bit)|

### Before downloading the docker image, please install Python27 or higher on your system:
  - https://www.python.org/downloads/release/python-latest

### Once installed, verify the python version:
      $ python -V
      Python 2.7.10
      
## Downloading and deploying Docker

### PLASMA
#### Download PLASMA Docker Image
```
docker pull wolkinc/plasma
```

#### Deploying PLASMA Docker Container
_Note: This command also automatically starts the PLASMA server_
```
docker run --name=plasma -dit --rm -p 5000:5000 -p 32003:32003 -p 31003:31003 wolkinc/plasma /Users/$USER/plasma/qdata/dd 32003 31003
```

## Verify PLASMA process

Once the Docker image and containers are deployed following above instructions, it will start the PLASMA process inside the Docker container. To verify if PLASMA is running:

##### From outside the Docker container
    $ docker exec `docker ps -a | grep plasma | awk '{print$1}'` ps aux | grep "plasma --datadir"
    
##### From within the Docker container
    $ ps aux | grep "plasma --datadir"
    
### PLASMA Port Mapping:

| Ports | Descriptions |
|--|--|
| 5001:5000 | <Syslog_system_port>:<Syslog_container_port> |
| 32003:32003 | <RPC_system_port>:<RPC_container_port> |
| 31003:31003 | <Network_Listening_system_port>:<Network_Listening_container_port> |

### SQL
#### Download SQL Docker Image
```
docker pull wolkinc/sql
```

#### Deploying SQL Docker Container
_Note: This command also automatically starts the SQL server_
```
docker run --name=sql -dit --rm -p 5001:5000 -p 24000:24000 -p 21000:21000 wolkinc/sql /Users/sql/qdata/dd 0xff6a2ac8b193b705 24000 21000 32003
```

##### OR the following command, if you want to specify your own ports:

```
docker run --name=sql -dit -p 5001:5000 -p RPCPORT:RPCPORT -p HTTPPORT:HTTPPORT wolkinc/sql /Users/$USER/sql/qdata/dd 0x3b6a2ac8b193b705 RPCPORT HTTPPORT PLASMAPORT

```

## Verify SQL process

Once the Docker image and containers are deployed following above instructions, it will start the SQL process inside the Docker container. To verify if SQL is running:

##### From outside the Docker container
    $ docker exec `docker ps -a | grep sql | awk '{print$1}'` ps aux | grep "sql --datadir"
    
##### From within the Docker container
    $ ps aux | grep "sql --datadir"
    
### SQL Port Mapping:

| Ports | Descriptions |
|--|--|
| 5001:5000 | <Syslog_system_port>:<Syslog_container_port> |
| 24000:24000 | <RPC_system_port>:<RPC_container_port> |
| 21000:21000 | <Network_Listening_system_port>:<Network_Listening_container_port> |

### NoSQL

#### Download NoSQL Docker Image
```
docker pull wolkinc/nosql
```
#### Deploying NOSQL Docker Container
_Note: This command also automatically starts the NoSQL server_

```
docker run --name=nosql -dit --rm -p 5001:5000 -p 34000:34000 -p 31000:31000 wolkinc/nosql /Users/$USER/nosql/qdata/dd 0x39f70e2d18affa78 34000 31000 32003
```

## Verify NoSQL process

Once the Docker image and containers are deployed following above instructions, it will start the NoSQL process inside the Docker container. To verify if NoSQL is running:

##### From outside the Docker container
    $ docker exec `docker ps -a | grep nosql | awk '{print$1}'` ps aux | grep "nosql --datadir"
    
##### From within the Docker container
    $ ps aux | grep "nosql --datadir"

### NoSQL Port Mapping:

| Ports | Descriptions |
|--|--|
| 5001:5000 | <Syslog_system_port>:<Syslog_container_port> |
| 34000:34000 | <RPC_system_port>:<RPC_container_port> |
| 31000:31000 | <Network_Listening_system_port>:<Network_Listening_container_port> |

### Attach/Detach/re-attach Docker container

#### In order to attach to the Docker Container shell:
      $ docker exec -it `docker ps -q` /bin/bash

#### In order to exit the Docker Container shell:
      $ ctrl + d (or simply type exit and Enter)

### To clean/delete all containers:
      $ docker stop `docker ps -a -q`;docker rm `docker ps -a -q`

### To clean/delete all the images (once the container/s have been deleted):
      $ docker rmi `docker images -q`
      
*_________________________________________________________________________________________________________________________*

## Docker COMMIT and PUSH to Docker Hub

### Repo required:
```git clone git@github.com:wolkdb/docker.git```

#### Login Credentials required to be able to push to docker hub:
* Username: anand@wolk.com
* Password: 14All41!

```
docker login --username anandray
Password:
```

### SQL

```
1. Copy new binary to "/root/go/src/github.com/wolkdb/docker/docker-sql/sql/bin/sql"

2. Create docker IMAGE: "docker build -t sql"

3. Create docker CONTAINER: "docker run --name=sql -dit --rm -p 5000:5000 -p 24000:24000 -p 21000:21000 sql /root/sql/qdata/dd 0xab6a2ac8b193b705"

4. Enter into the SQL CONTAINER to make sure its running or run RPC query from outside to verify: 
"docker exec -it `docker ps -a -q` /bin/bash"

5. Record docker CONTAINER ID: "docker ps"

6. Commit the image using CONTAINER_ID from STEP 5: "docker commit -a "email@wolk.com" CONTAINER_ID sql"

7. Run "docker images" to record docker IMAGE ID

8. Add tag using TAG_ID from STEP 7: "docker tag TAG_ID wolkinc/sql"

9. Pushing to docker hub: "docker push wolkinc/sql"
```

### NOSQL

```
1. Copy new binary to "/root/go/src/github.com/wolkdb/docker/docker-nosql/nosql/bin/nosql"

2. Create docker IMAGE: "docker build -t nosql"

3. Create docker CONTAINER: "docker run --name=nosql -dit --rm -p 5000:5000 -p 34000:34000 -p 31000:31000 nosql /root/nosql/qdata/dd 279"

4. Enter into the NOSQL CONTAINER to make sure its running or run RPC query from outside to verify: 
"docker exec -it `docker ps -a -q` /bin/bash"

5. Record docker CONTAINER ID: "docker ps"

6. Commit the image using CONTAINER_ID from STEP 5: "docker commit -a "email@wolk.com" CONTAINER_ID nosql"

7. Run "docker images" to record docker IMAGE ID

8. Add tag using TAG_ID from STEP 7: "docker tag TAG_ID wolkinc/nosql"

9. Pushing to docker hub: "docker push wolkinc/nosql"
```
