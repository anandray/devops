#!/usr/bin/php
<?php
include "storage.php";
error_reporting(E_ERROR);

function myexec($cmd, $run) {
    if ( $run ) {
      $output = array();
      exec($cmd, $output);
      if ( count($output) > 0 ) {
          return($output[0]);
      }
    }
}

getWolkDatabase(true);

$sql = "select hostname, publicip, datacenter, cloudprovider, consensus, nodenumber from servers where pushwolk=1 order by nodenumber, hostname";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $servers[] = $a;
  }
} else {
  echo mysql_error();
  exit(0);
}
$stage = isset($argv[1]) ? $argv[1] : 0;
$name = isset($argv[2]) ? $argv[2] : "sourabh";

if ( $stage == 0 )  $stage = "";
$httpport = 80 + $stage;
$maindir = "/root/go/src/github.com/wolkdb/cloudstore";
$run = true;
$binary = "$maindir/build/bin/wolk$stage";
$cmd0 = "ssh c0.wolk.com 'md5sum $binary'";
$md5sum = myexec($cmd0, $run);
foreach ($servers as $i => $s) {
  $hostname = $s->hostname;
  $server = $s->publicip;
  $binary = "$maindir/build/bin/wolk$stage";
  $md5sum2 = myexec("ssh -q $server 'md5sum $binary'", $run);
  if ( $md5sum2 == $md5sum ) {
    echo "PASS ";
    $passes++;
  } else {
    echo "FAIL ";
  }
  echo "$hostname\t$md5sum2\n";
}
echo "$passes/".count($servers)."\n";
?>
