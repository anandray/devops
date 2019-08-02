<?
include_once "storage.php";
include_once "dataproc.php";
include_once "communications.php";

define(SERVICE_ACCOUNT, "get_inventory_servers");

function setup_nagios($servers)
{
  $cmds = array();
  $cmds[] = "rm -f /usr/local/nagios/etc/objects/servers/www2*.cfg";
  $cmds[] = "rm -f /usr/local/nagios/etc/objects/servers/www6*.cfg";
  $cmds[] = "rm -f /usr/local/nagios/etc/objects/servers/www8*.cfg";
  $cmds[] = "rm -f /usr/local/nagios/etc/objects/servers/wwwa*.cfg";
  $cmds[] = "sh /var/www/vhosts/mdotm.com/hadoop/systems/dataproc/remove_dead_servers.sh";
  $cmds[] = "chown -R nagios.nagcmd /usr/local/nagios/share/";
  $cmds[] = "chown -R nagios.nagcmd /usr/local/nagios/etc/objects/";
  foreach ($cmds as $cmd) {
    echo "$cmd\n";
    exec($cmd, $a);
  }
  foreach ($servers as $hostgroup => $hosts) {
    foreach ($hosts as $hostname => $ip) {
      $fn = "/usr/local/nagios/etc/objects/servers/$hostname.cfg";
      $str = "define host{
        host_name               $hostname
        address                 $hostname
        contact_groups          oncall-admins,oncall-admins1
        check_command           check-host-alive
        max_check_attempts      10
        notification_interval   120
        notification_period     24x7
        notification_options    d,u,r
    }";
      if ( $f  = fopen($fn, "w") ) {
	fputs($f, "$str\n");
	fclose($f);
      }
    }
  }
  fclose($f);
  $cmds = array();
  $cmds[] = "ssh www6002 service nagios reload";
  foreach ($cmds as $cmd) {
    exec($cmd, $a);
  }
  return(true);
}


$storage = new Storage;
$storage->getMdotmDatabase();

//$sql = "select hostname, publicip, privateip from servers where (hostname like 'www2%' or hostname like 'www6%' or hostname like 'wwwa%' or hostname like 'www8%') and createDT < date_sub(Now(), interval 10 minute)";
//$sql = "select hostname, publicip, privateip from servers where (hostname like 'www2%' or hostname like 'www6%' or hostname like 'wwwa%' or hostname like 'www8%' or hostname like 'ha6%') and createDT < date_sub(Now(), interval 10 minute)";
$sql = "select hostname, publicip, privateip from servers where (hostname like 'www2%' or hostname like 'www6%' or hostname like 'wwwa%' or hostname like 'www8%' or hostname like 'ha6%')";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $hostname = $a->hostname;
    $ip = $a->publicip;
    $hostgroup = substr($hostname, 0, 4);
    if ( strlen($hostname) > strlen("www6001") ) {
      switch ($hostgroup) {
      case "www2":
      case "www6":
      case "www8":
      case "wwwa":
      case "ha6":
	$servers[$hostgroup][$hostname] = $ip;
      break;
      }
    }
  }
}



if ( setup_nagios($servers) ) {
  gc_success(SERVICE_ACCOUNT);
}
?>
