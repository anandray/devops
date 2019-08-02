<?php
include "storage.php";
error_reporting(E_WARNING);

getWolkDatabase(true);

$sql = "select hostname, datacenter, publicip, cloudprovider, nodenumber, consensus from servers order by consensus, nodenumber";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $hostname = $a->hostname;
    $ip = $a->publicip;
    if ( $a->consensus == 1 ) {
        $nodenumber = $a->nodenumber;
        $dnsname = "c$nodenumber.wolk.com";

      // do port 443 checks, healthchecks 
      $template = "define host {
         host_name                       consensus-$hostname
         address                         $dnsname
         contact_groups                  admins
         check_command                   check-host-alive
         max_check_attempts              10
         notification_interval           120
         notification_period             24x7
         notification_options            d,u,r
         }";
    } else {
      // do port 80 checks + healthchecks
      $template = "define host {
         host_name                       storage-$hostname
         address                         $ip
	 contact_groups                  admins
         check_command                   check-host-alive
         max_check_attempts              10
         notification_interval           120
         notification_period             24x7
	 notification_options            d,u,r
	 }";
    }
    echo "$template\n";
  }
} else {
  echo mysql_error();
  exit(0);
}

echo "\n\n";

?>
