#!/bin/bash
/usr/local/share/google/google-cloud-sdk/bin/gsutil -m acl ch -f -R -u AllUsers:R gs://crosschannel_cdn
