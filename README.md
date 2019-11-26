# Wolk Dev Ops

## Local Development 
* [Wolk Local Development - Go, Github Desktop and Atom](https://docs.google.com/document/d/1LRcjYV_Qu0-c5g3p_ka1WqhYY48-bRjss2RgP5GSoZc/edit)

## GoLang
* [Wolk Golang notes](https://docs.google.com/document/d/1RACPpdKeq3SAMqB1a4-z6qFRPvbfWbnw2yzYIR4WcwA/edit#heading=h.lavyprn7v642)

## Git
* [Submodules cheatsheet](https://docs.google.com/document/d/1z6df3Xxmf_rXTwkODLHg6G4P4LR7nocRoymAZXUKwUA/edit)

## Kubernetes / Docker
* [Docker / Kubernetes Cheat Sheet](https://docs.google.com/document/d/1ZsQ3_WgvIHf92e2sPcVXuEfseoQ3mZISdmKQ2xiF9jU/edit)
* [Kubernetes across Cloud Providers](https://docs.google.com/document/d/14lwSSygwcL5NJecyTUbLq1YjnyxxfrStxO4UBQPFuB8/edit)
* [Kubernetes - Add nodes to existing cluster](https://docs.google.com/document/d/1ZGTdU_imxuOoahUHP8IGwILl7S263EJn9v5ywWQ-sro/edit?usp=sharing)
* [Kubernetes (Alibaba)](https://docs.google.com/document/d/1TTKVWmfOT1de-5O7BgstwvjXrxOLiA7qxo_sV9LTR5s/edit)
* [KubeMCI](https://docs.google.com/document/d/1tmjQmBZoHgYPJT2vg7Jc84Kkf5BJm5pA7Lvu1SOhNBI/edit?usp=sharing)

## Ceph
* [Ceph OSD Add/Removal](ttps://docs.google.com/document/d/1u-WQXbYQnDVJpHluBRAUeyzEDOAe_WQJ3H3XYpPwr_M/edit)
* [ceph-repair/recovery](https://docs.google.com/document/d/1XB2nTf8bXn8o8adbk8wBUacByvbor5Ppnm2l7QjgpSE/edit)
* Ceph Dashboards: [cephy](http://cephy2.wolk.com:7000/) [cephz](http://dash.wolk.com:7000/health)
* Graphana Dashboards: https://computingforgeeks.com/monitoring-ceph-cluster-with-prometheus-and-grafana/ 
* Ceph on Mac: https://docs.google.com/document/d/16NXWAV6ukiAw3kYvytjHfqp1GUtwTTCfMsLyMtS5dEY/edit
* Create replicapool:
  ```
  for i in {0..15}; do ceph osd pool create x$i 10; done
  for i in {0..15}; do ceph osd pool application enable x$i rbd; done
  ```

## Cloudflare
* [Cloudflare DNS API](https://docs.google.com/spreadsheets/d/1Dsz4lbtwjAoTY2WHoh30eydhBxT9akCYf1BNwVXtzMw/edit#gid=0)

## GCE
* [Whitelist IP in GCE Firewall](https://docs.google.com/document/d/1_Ovwv2gK0_E3ctioPBg5cYpyKpATjFMB5jyhnUeUafU/edit)

## Handshake
* [Handshake TLSd](https://docs.google.com/spreadsheets/d/1yhnHs-ARUTUPqSFlwXbyJfer6ZOhbN0F0aJBHXM77XQ/edit#gid=1186001804)

## Monitoring
* [testnet](https://testnet.wolk.com)
* [cacti](https://cacti.wolk.com/) admin/14All41$
* [nagios](https://nagios.wolk.com/nagios/)  admin/14All41$

## Goals
* [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/2144381)
* [Quarterly Goals](https://docs.google.com/document/d/1fllGqnDC1HS5N85amvGkW7HXTtvncRQvW3B7ixkJqiU/edit)
