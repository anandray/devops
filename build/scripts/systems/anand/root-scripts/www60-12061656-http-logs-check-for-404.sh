grep " 404 " /var/log/httpd/mdotm.com-access_log  | grep -v healthcheck | head -n10
