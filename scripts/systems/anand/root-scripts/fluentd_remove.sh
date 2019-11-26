#!/bin/bash

# stop/remove google-fluentd
sudo /sbin/service google-fluentd stop;
sudo yum -y remove google-fluentd google-fluentd-catch-all-config;
