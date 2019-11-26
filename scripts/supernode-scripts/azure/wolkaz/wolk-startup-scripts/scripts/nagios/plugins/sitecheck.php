#!/usr/bin/php
<?
error_reporting(E_WARNING);
$check = "site";
$host = $argv[1];
$url = "http://".$host."/ads/systems/check.php?check=".$check;
$res = file_get_contents($url);

if ( $f = fopen("/var/log/check.log", "a+") ) {
  $tm = date("H:i:s", time());
  fwrite($f, "$tm|$host|$url|$res|$check\n");
  fclose($f);
 }

$sa = explode(":", $res);
if ( count($sa) == 2 ) {
  $code = intval($sa[0]);
  $msg = $sa[1];
  switch ( $code ) {
  case 2:
    echo "CRITICAL - $msg: $host \n";
    exit(2);
  case 1:
    echo "WARNING - $msg: $host \n";
    exit(1);
  default:
  case 0:
    echo "OK - $msg: $host \n";
    exit(0);
  }
 } else {
  echo "UNKNOWN\n";
  exit(3);
 }
echo "DEFAULT response for $check";
exit(0);
?>
