<?php

function put_wcloud($p, $owner, $collection) {
  $hp = "-httpport=$p";
  
  // setname
  $cmd = "wcloud $hp setname $owner";
  echo "$cmd\n";
  $res = exec($cmd, $a);
  
  // getname
  $done = false;
  while ( ! $done ) {
    $cmd = "wcloud $hp getname $owner";
    $a = array();
    $res = exec($cmd, $a, $r);
    if ( $res == "545df6fd6811dc32397eae8d218fc4b29681eb62" ) { 
      // mkdir 
      $cmd = "wcloud $hp mkdir wolk://$owner/$collection";
      $res = exec($cmd, $a);
      // TODO: check if mkdir successful
      $done = true;
      sleep(2);
      if ( $tries++ > 100 ) {
	exit(0);
      }
    } else {
      sleep(1);
    }
  }
  
  $arr = array("wfs.html"  => "36406c635fe39a678435bea908e5b59f",
	       "banana.gif" => "77a0a982809ebe6b898bbed8fe0a0013",
	       "pot.jpeg" => "67a74bedc47c1d9d68d1f25b3fd1e0e0",
	       "zone.txt" => "ed1c907f6759d9372069e4742bd9bae2",
	       "cat.jpeg" => "b84d6366deec053ff3fa77df01a54464", 
	       "SampleVideo_1280x720_1mb.mp4" => "d55bddf8d62910879ed9f605522149a8"
	       );
  
  $server = "c0.wolk.com";
  foreach ($arr as $fn => $md5) {
    $fullfn = "/root/go/src/github.com/wolkdb/cloudstore/content/$fn";
    $cmd = "wcloud -server=$server $hp put $fullfn wolk://$owner/$collection/$fn";
    echo "$cmd\n";
    $url = "wolk://$owner/$collection/$fn";
    exec($cmd, $s);
    echo "URL: $url\n";
  }
  
  sleep(2);
  foreach ($arr as $fn => $md5) {
    $url = "wolk://$owner/$collection/$fn";
    $cmd = "wcloud -server=$server $hp get $url | md5sum";
    echo "$cmd\n";
    $s = array();
    $res = exec($cmd, $s);
    $res = trim($res, " -\n");
    $md5 = trim($res, " \n");
    if ( $res != $md5 ) {
      echo $res."=?=".$md5." FAILURE\n";
    } else {
      echo "SUCC\n";
    }
  }
}

$p = isset($argv[1]) ? $argv[1] : 443;
$owner = isset($argv[2]) ? $argv[2] : "bruce";
$collection = isset($argv[3]) ? $argv[3] : "test";
put_wcloud($p, $owner, $collection);
?>