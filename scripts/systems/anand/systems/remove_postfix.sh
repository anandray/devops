#!/bin/bash
service postfix stop ; rpm -e postfix --nodeps ; rm -rf /var/spool/postfix ; userdel postfix
