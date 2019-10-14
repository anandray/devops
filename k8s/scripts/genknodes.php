#!/usr/bin/php
<?php
include "storage.php";
error_reporting(E_ERROR);

$mysqli = getWolkDatabase(true);
$sql = "select * from knodes order by nodenumber";
if ( $res = $mysqli->query($sql) ) {
  while ( $a = $res->fetch_object() ) {
    $knodes[] = $a;
  }
} else {
  echo $mysqli->error();
}

$knodesstr = "var registry=".json_encode($knodes).";\n";
file_put_contents("/root/go/src/github.com/wolkdb/devops/testnet/knodes.js", $knodesstr);
echo "$knodesstr\n";


?>
