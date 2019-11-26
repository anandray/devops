<?php
include_once "communications.php";
include_once "storage.php";

/*
echo "\n------------------------------\n";
echo "return: $return";
echo "\n------------------------------\n";
print_r($output);
echo "\n------------------------------\n";		
echo "return_status: $return_status";		
echo "\n------------------------------\n";	
exit(0);
*/

$hostname = php_uname('n');
$res = "";

// STEP 1
// sudo + gsutil
//$sudo_cmd = "/usr/bin/sudo /usr/bin/ssh root@`hostname` gsutil ls  2>/dev/null"; 
//$sudo_cmd = "gsutil ls  2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$sudo_cmd = "ssh `hostname` sudo gsutil ls 2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$return = "";
$output = array();
$return_status = -1;
$return = exec($sudo_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "gs://")) ){
	$res .= "STEP 1 - sudo ok  |  \n";
} else {	
	$res .= "STEP 1 - sudo fail\n";
	return new_server_check_fail($res);
}

// STEP 2
// hosts
$database = array("db01", "db02", "db03", "db4", "db04", "dbc3");
foreach ($database as $db) {
  $hosts_cmd = "mysql -A -udb -p1wasb0rn2 -h$db -e 'show databases'";
  $return = "";
  $output = array();
  $return_status = -1;
  $return = exec($hosts_cmd, $output, $return_status);  
  if ( ($return_status == 0) && (count($output)>1) ){
  	  $res .= "STEP 2 - hosts $db ok  |  \n";
  } else {
  	  $res .= "STEP 2 - hosts $db fail\n";
	  return new_server_check_fail($res);
  }
}
 
// STEP 3
// ssh
$ssh_cmd_1 = "ssh -q www6002 hostname 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($ssh_cmd_1, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "www6002")) ){
	$res .= "STEP 3 - ssh (1) ok  |  \n";
} else {	
	$res .= "STEP 3 - ssh (1) fail\n";
	return new_server_check_fail($res);
}

$ssh_cmd_2 = "ssh -q `hostname` id 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($ssh_cmd_2, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "uid=0(root) gid=0(root)")) ){
	$res .= "STEP 3 - ssh (2) ok  |  \n";
} else {	
	$res .= "STEP 3 - ssh (2) fail\n";
	return new_server_check_fail($res);
} 

// STEP 4
// SELINUX
$selinux_cmd = "curl -I http://`hostname`/inf0.php 2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$return = "";
$output = array();
$return_status = -1;
$return = exec($selinux_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "200 OK")) ){
    $res .= "STEP 4 - selinux ok  |  \n";
} else {
	$res .= "STEP 4 - selinux fail\n";
	return new_server_check_fail($res);
} 		
 
// STEP 5
// ulimit
$selinux_cmd = "ulimit -a | egrep 'open files|processes' 2>/dev/null"; // ulimit -n  and  ulimit -u
$return = "";
$output = array();
$return_status = -1;
$return = exec($selinux_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "500000")) && (stristr($output[1], "500000")) ){
	$res .= "STEP 5 - ulimit ok  |  \n";
} else {	
	$res .= "STEP 5 - ulimit fail\n";
	return new_server_check_fail($res);
} 		

// STEP 6
// PHP
$php_cmd_1 = "php -m | egrep -i 'lua|pdf|apc|geo|max|igbinary|ssh2' | /usr/bin/wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($php_cmd_1, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "7")) ){
	$res .= "STEP 6 - php (1) ok  |  \n";
} else {	
	$res .= "STEP 6 - php (1) fail\n";
	return new_server_check_fail($res);
}

$php_cmd_2 = "php -m | /usr/bin/wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($php_cmd_2, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "65")) ){
	$res .= "STEP 6 - php (2) ok  |  \n";
} else {
	$res .= "STEP 6 - php (2) fail\n";
	return new_server_check_fail($res);
}

// STEP 7
// java
$java_cmd = "sudo java -version 2>&1";
$return = "";
$output = array();
$return_status = -1;
$return = exec($java_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "1.8")) ){
	$res .= "STEP 7 - java ok  |  \n";
} else {
	$res .= "STEP 7 - java fail\n";
	return new_server_check_fail($res);
}

// STEP 8
// syslog
$syslog_cmd = "sudo service syslog status 2>&1";
$return = "";
$output = array();
$return_status = -1;
$return = exec($syslog_cmd, $output, $return_status);
if ( ($return_status == 1) && (stristr($output[0], "unrecognized")) ){
	$res .= "STEP 8 - syslog ok  |  \n";
} else {
	$res .= "STEP 8 - syslog fail\n";
	return new_server_check_fail($res);
}

// STEP 9
// syslog-ng
$syslog_ng_cmd_1 = "sudo service syslog-ng status 2>&1";
$return = "";
$output = array();
$return_status = -1;
$return = exec($syslog_ng_cmd_1, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "running")) ){
	$res .= "STEP 9 - syslog-ng (1) ok  |  \n";
} else {
	$res .= "STEP 9 - syslog-ng (1) fail\n";
	return new_server_check_fail($res);
}

$json = array(		
		"machine" => (string)$hostname		
);
// {"machine":"www61-08120419-017p"}
// ssh -q log6 cat /disk1/log/log/2016/08/18/162*.log | grep '"machine":"www61-08120419-017p"'
$stream = 'cc-probe';
openlog($stream, LOG_PID | LOG_ODELAY, LOG_LOCAL4); syslog(LOG_INFO, json_encode($json, JSON_UNESCAPED_SLASHES));  closelog(); 
$logDT = date("Y/m/d/Hi", time()); 
//$syslog_ng_cmd_2 = "ssh -q log6 cat /disk1/log/log/2016/08/17/1714.log | grep '\"machine\":\"$hostname\"' | wc -l"; 
$syslog_ng_cmd_2 = "ssh -q log6 cat /disk1/log/probe/$logDT.probe | grep '\"machine\":\"$hostname\"' | wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($syslog_ng_cmd_2, $output, $return_status);
if ( ($return_status == 0) && ($output[0]>0) ){
	$res .= "STEP 9 - syslog-ng (2) ok  |  \n";	
} else {
	$res .= "STEP 9 - syslog-ng (2) fail\n";
	return new_server_check_fail($res);
}

// STEP 10
// postfix
$postfix_cmd = "ps aux | grep postfix | grep -v grep | wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($postfix_cmd, $output, $return_status);
if ( ($return_status == 0) && ($output[0] == 0) ){
	$res .= "STEP 10 - postfix ok  |  \n";
} else {
	$res .= "STEP 10 - postfix fail\n";
	return new_server_check_fail($res);
}

// STEP 11
// sendmail
$sendmail_cmd = "ps aux | grep sendmail | grep -v grep | grep clientmqueue | wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($sendmail_cmd, $output, $return_status);
if ( ($return_status == 0) && ($output[0] == 1) ){
	$res .= "STEP 11 - sendmail ok  |  \n";
} else {
	$res .= "STEP 11 - sendmail fail\n";
	return new_server_check_fail($res);
}

// STEP 12
// php_short_open_tag   check for short_open_tag = On   <?    
$php_short_open_tag_cmd = "grep \"short_open_tag = On\" /etc/php.ini | wc -l 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($php_short_open_tag_cmd, $output, $return_status);
if ( ($return_status == 0) && ($output[0] == 1) ){
	$res .= "STEP 12 - php_short_open_tag ok  |  \n";
} else {
	$res .= "STEP 12 - php_short_open_tag fail\n";
	return new_server_check_fail($res);
}

// STEP 13
// hbase
$hbase_cmd = "curl \"http://$hostname/ads/systems/hbasehealthcheck.php\" 2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$return = "";
$output = array();
$return_status = -1;
$return = exec($hbase_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "OK")) ){
	$res .= "STEP 13 - hbase ok  |  \n";
} else {
	$res .= "STEP 13 - hbase fail\n";
	return new_server_check_fail($res);
}

// STEP 14
// geo
$geo_cmd = "curl \"http://`hostname`/ads/systems/check.php?check=geo\" 2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$return = "";
$output = array();
$return_status = -1;
$return = exec($geo_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "OK")) ){
	$res .= "STEP 14 - geo ok  |  \n";
} else {
	$res .= "STEP 14 - geo fail\n";
	return new_server_check_fail($res);
}

// STEP 15
// maxmind
include_once 'storage.php';
require_once 'vendor/autoload.php';
use GeoIp2\Database\Reader;
$ipaddress = "8.8.8.8 ";
$reader = new Reader('/usr/share/GeoIP/GeoIP2-City.mmdb');
$record = $reader->city($ipaddress);
$return = "";
$output = array();
$return_status = -1;
if ( stristr($record->city->names[en], "Mountain View") ){
	$res .= "STEP 15 - maxmind ok  |  \n";
} else {
	$res .= "STEP 15 - maxmind fail\n";
	return new_server_check_fail($res);
}

// STEP 16
// timezone
$date = new DateTime();
$tz = $date->getTimezone();
$tz_name = $tz->getName();
$return = "";
$output = array();
$return_status = -1;
if ( stristr($tz_name, "America/Los_Angeles") ){
	$res .= "STEP 16 - timezone ok  |  \n";
} else {
	$res .= "STEP 16 - timezone fail\n";
	return new_server_check_fail($res);
}

// STEP 17
// apc
$apc_cmd = "curl \"http://`hostname`/ads/systems/apc.php\" 2>/dev/null"; // redirect stderr to /dev/null  (0 is stdin. 1 is stdout. 2 is stderr)
$return = "";
$output = array();
$return_status = -1;
$return = exec($apc_cmd, $output, $return_status);     // retern -> string(3) "BAR"
if ( ($return_status == 0) && (stristr($output[0], "BAR")) ){
	$res .= "STEP 17 - apc ok  |  \n";
} else {
	$res .= "STEP 17 - apc fail\n";
	return new_server_check_fail($res);
}

// STEP 18
// crond
//$crond_cmd = "/sbin/service crond status 2>/dev/null";
$crond_cmd = "sudo service crond status 2>/dev/null";
$return = "";
$output = array();
$return_status = -1;
$return = exec($crond_cmd, $output, $return_status);
if ( ($return_status == 0) && (stristr($output[0], "running")) ){
	$res .= "STEP 18 - crond ok  |  \n";
} else {
	$res .= "STEP 18 - crond fail\n";
	return new_server_check_fail($res);
}

// STEP 19
// shortcircuit
include_once 'shortcircuit.php';
$return = "";
$output = array();
$return_status = -1;
if ( isset($shortcircuit) ){
	$res .= "STEP 19 - shortcircuit ok  |  \n";
} else {
	$res .= "STEP 19 - shortcircuit fail\n";
	return new_server_check_fail($res);
}

function new_server_check_fail($res){
  echo $res;
	$debug_str = str_replace(array("\n", ":", "'"), array("", "", ""), $res);
	debuglog("new_server_check:FAIL:$debug_str");
  gc_error("new_server_check", "new_server_check Error", $res, NEW_SERVER_ERROR_CODE, "", 0, "", "2");
  return 0;
}

echo $res;
debuglog("new_server_check:OK");
gc_success("new_server_check");
return 1;






/*
 
echo "\n------------------------------\n";
echo "return: $return";
echo "\n------------------------------\n";
print_r($output);
echo "\n------------------------------\n";		
echo "return_status: $return_status";		
echo "\n------------------------------\n";	
exit(0); 
 





 //lua
//igbinary
// mail 


// pdf  http://www.java2s.com/Code/Php/PDF/HelloWorldUsingPDFLib.htm
root@anand-test-1-mhvi systems]# curl -I http://`hostname`/ads/systems/pdf.php
HTTP/1.1 200 OK
Date: Tue, 16 Aug 2016 00:46:05 GMT
Server: Apache/2.2.15 (CentOS)
X-Powered-By: PHP/5.4.45
Content-disposition: inline; filename=mypdf.pdf
Content-Length: 2734
Connection: close
Content-Type: application/pdf



// httpd_conf
#Add LogFormat + vhosts + etc...

*/


?>
