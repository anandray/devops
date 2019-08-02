#!/bin/bash
mkdir -p /var/www/.config/gcloud /var/www/.gsutil
chgrp -R -v apache /var/www/.config/gcloud /var/www/.gsutil
chmod -v 0775 /var/www/.config/gcloud /var/www/.gsutil
