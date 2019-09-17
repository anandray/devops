#!/usr/bin/php
<?
$host = $argv[1];
//$warn_node = $argv[2]; #WARNING when there is less than this number of nodes alive
//$crit_node = $argv[3]; #CRITICAL when there is less than this number of nodes alive
if ( $host == "10.16.100.252" || ( $host == "10.16.100.230" ) ) {
  $url = "http://".$host.":8080/ads/appcheck.php";
 } else {
  $url = "http://".$host."/ads/appcheck.php";
 }
$res = file_get_contents($url);

if ( $f = fopen("/var/log/appcheck.log", "a+") ) {
  $tm = date("H:i:s", time());
  fwrite($f, "$tm|$host|$url|$res\n");
  fclose($f);
 }
$sa = explode(":", $res);
if ( count($sa) == 2 ) {
  $code = intval($sa[0]);
  $msg = $sa[1];
  
  switch ( $code ) {
  case 2:
    echo "CRITICAL - $msg: $host \n";
//    echo "APPCHECK CRITICAL \n";
    exit(2);
  case 1:
    echo "WARNING - $msg: $host \n";
//    echo "APPCHECK WARNING \n";
    exit(1);
  default:
  case 0:
    echo "OK - $msg: $host \n";
//    echo "APPCHECK OK \n";
    exit(0);
  }
 } else {
  echo "UNKNOWN\n";
  exit(3);
 }
?>
