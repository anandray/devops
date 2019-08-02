#!/bin/bash

# Downloading ORACLE Data using the following LFTP command and credentials:
#lftp <<EOF
#open -u mdotm_sftp,Qs5fo/nWtIAy sftp://batcher.bluekai.com
#mget *_`date +%Y%m%d`.log.gz
#bye
#EOF
# /End of command

# Cronjob at 1AM
#0 1 * * * sh /var/www/vhosts/mdotm.com/scripts/systems/oracle_download.sh &> /var/log/oracle.log

# The activity is logged in /var/log/oracle.log

if [ ! -d /disk1/oracle/oracle_`date +%m%d%Y` ];
  then
    echo "[1] Creating directory /disk1/oracle/oracle_`date +%m%d%Y`"
    mkdir -p /disk1/oracle/oracle_`date +%m%d%Y`
  else
    echo "[1] /disk1/oracle/oracle_`date +%m%d%Y` EXISTS"
fi

if [ ! -f /disk1/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz ];
  then
    cd /disk1/oracle/oracle_`date +%m%d%Y`
    echo "[2] Downloading /disk1/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz"
    lftp <<EOF
    open -u mdotm_sftp,Qs5fo/nWtIAy sftp://batcher.bluekai.com
    mget *_`date +%Y%m%d`.log.gz
    bye
EOF
  else
    echo "[2] /disk1/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz EXISTS, no need to download"
fi

if [ ! -f /disk1/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz.bz2 ];
  then
    echo "[3] Bzipping today's downloaded data"
    ls -1 /disk1/oracle/oracle_`date +%m%d%Y` | grep .gz | awk '{print"zcat /disk1/oracle/oracle_`date +%m%d%Y/`"$1,"| bzip2 -v > /disk1/oracle/oracle_`date +%m%d%Y`/"$1".bz2"}' | grep `date +%Y%m%d` > /usr/bin/oracle_bz2.sh
    cd /disk1/oracle/oracle_`date +%m%d%Y`
    sh /usr/bin/oracle_bz2.sh
  else
    echo "[3] /disk1/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz.bz2 EXISTS, no need to bzip"
fi

# copy entire directory /disk1/oracle/oracle_`date +%m%d%Y` to gs://crosschannel_segments/oracle/

if ! /usr/local/share/google/google-cloud-sdk/bin/gsutil ls gs://crosschannel_segments/oracle/oracle_`date +%m%d%Y`/mdotm_sftp_`date +%Y%m%d`.log.gz.bz2 > /dev/null;
  then
    echo "[4] Copying today's data to Google Storage:"
    /usr/local/share/google/google-cloud-sdk/bin/gsutil -m cp -r /disk1/oracle/oracle_`date +%m%d%Y` gs://crosschannel_segments/oracle/
  else
    echo "[4] gs://crosschannel_segments/oracle/oracle_`date +%m%d%Y` EXISTS, no need to copy"
fi

# look for data from SEVEN days back in Google Storage and delete from www6002 if it exists in gs://crosschannel_segments/oracle/oracle_`date +%m%d%Y --date='7 days ago'`

if gsutil ls gs://crosschannel_segments/oracle/oracle_`date +%m%d%Y --date='7 days ago'`/mdotm_sftp_`date +%Y%m%d --date='7 days ago'`.log.gz.bz2 > /dev/null;
  then
    echo "[5a] SEVEN DAYS OLD DATA - oracle_`date +%m%d%Y --date='7 days ago'`/mdotm_sftp_`date +%Y%m%d --date='7 days ago'`.log.gz EXISTS in gs://crosschannel_segments/oracle/"
    echo "[5b] Deleting SEVEN DAYS OLD gz copy from www6002:/disk1/oracle/oracle_`date +%m%d%Y --date='7 days ago'`/mdotm_sftp_`date +%Y%m%d --date='7 days ago'`.log.gz"
    rm -rfv /disk1/oracle/oracle_`date +%m%d%Y --date='7 days ago'`
  else
    echo "[5a] oracle_`date +%m%d%Y --date='7 days ago'`/mdotm_sftp_`date +%Y%m%d --date='7 days ago'`.log.gz does NOT exist in gs://crosschannel_segments/oracle/"
    echo "[5b] Trying to copy again to gs://oracle/oracle_`date +%m%d%Y --date='7 days ago'`/"
    /usr/local/share/google/google-cloud-sdk/bin/gsutil -m cp -r /disk1/oracle/oracle_`date +%m%d%Y --date='7 days ago'` gs://crosschannel_segments/oracle/
fi
