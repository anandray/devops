#!/bin/bash
iptables -F;
rm -rf /etc/sysconfig/iptables.save;
iptables -A INPUT -s 10.84.81.165 -p tcp --dport 5666 -j ACCEPT; #www1016 private
iptables -A INPUT -s 75.126.67.187 -p tcp --dport 5666 -j ACCEPT; #www1016 public
iptables -A INPUT -s 10.84.51.50 -p tcp --dport 5666 -j ACCEPT;
iptables -A INPUT -s 0.0.0.0/0 -p tcp --dport 5666 -j DROP;
iptables -A INPUT -s 10.0.0.0/8  -p tcp --dport 22 -j ACCEPT; #Allow SSH access to and from within SoftLayer servers
iptables -A INPUT -s 10.84.81.165 -p tcp --dport 22 -j ACCEPT; #www1016 private
iptables -A INPUT -s 75.126.67.187  -p tcp --dport 22 -j ACCEPT; #www1016 public
iptables -A INPUT -s log04  -p tcp --dport 22 -j ACCEPT;
iptables -A INPUT -s www1001  -p tcp --dport 22 -j ACCEPT;
iptables -A INPUT -s www1011  -p tcp --dport 22 -j ACCEPT; #www1016 public
iptables -A INPUT -s admin  -p tcp --dport 22 -j ACCEPT; #www1016 public
iptables -A INPUT -s ha00  -p tcp --dport 22 -j ACCEPT;
iptables -A INPUT -s 75.126.21.8  -p tcp --dport 22 -j ACCEPT; #HA00 public
iptables -A INPUT -s 127.0.0.1 -p tcp --dport 111 -j ACCEPT;
iptables -A INPUT -s 127.0.0.1 -p udp --dport 111 -j ACCEPT;
sh /root/scripts/iptables_public.sh; #All Softlayer Public IPs
sh /root/scripts/iptables_public_portmap_tcp_111.sh; #All Softlayer Public IPs
sh /root/scripts/iptables_public_portmap_udp_111.sh; #All Softlayer Public IPs
iptables -A INPUT -s 50.225.47.189 -p tcp --dport 22 -j ACCEPT; #Office New Fiber
iptables -A INPUT -s 50.225.47.128/26 -p tcp --dport 22 -j ACCEPT; #Office New Fiber
iptables -A INPUT -s 67.180.166.145 -p tcp --dport 22 -j ACCEPT; #Office New Cable
ip6tables -A INPUT -s 2601:646:c200:9d00::/64 -p tcp --dport 22 -j ACCEPT; #Office New Cable IPv6
iptables -A INPUT -s 50.233.2.134 -p tcp --dport 22 -j ACCEPT; #Office_Fibre
iptables -A INPUT -s 24.5.3.39 -p tcp --dport 22 -j ACCEPT; #Office
iptables -A INPUT -s 24.5.130.231 -p tcp --dport 22 -j ACCEPT; #Sourabh
iptables -A INPUT -s 73.202.187.137 -p tcp --dport 22 -j ACCEPT; #anand
iptables -A INPUT -s 76.126.201.220 -p tcp --dport 22 -j ACCEPT; #anand
iptables -A INPUT -s 73.223.30.42 -p tcp --dport 22 -j ACCEPT; #Yaron
iptables -A INPUT -s 50.185.118.140 -p tcp --dport 22 -j ACCEPT; #Yaron_new
iptables -A INPUT -s 73.189.213.45 -p tcp --dport 22 -j ACCEPT; #Yaron_new
iptables -A INPUT -s 209.117.57.66 -p tcp --dport 22 -j ACCEPT; #Chicago
iptables -A INPUT -s 0.0.0.0/0 -p tcp --dport 22 -j DROP;
iptables -A INPUT -p tcp -m tcp --dport 5060 -j REJECT;
iptables -A INPUT -p udp -m udp --dport 5060 -j REJECT;
iptables -A INPUT -p udp -m udp --dport 53 -j REJECT
iptables -A INPUT -p tcp -m tcp --dport 53 -j REJECT
iptables -A INPUT -p tcp -m tcp --dport 111 -j REJECT
#iptables -A INPUT -p tcp -m udp --dport 111 -j REJECT
service iptables save;
sed -i '/iptables/d' /etc/rc.local &&
sed -i "$ i\/bin/sh /root/scripts/iptables.sh" /etc/rc.local
