<?php
include "storage.php";

$bin = "/root/go/src/github.com/wolkdb/wolkjs/wcloud";

$port = 81;
$n = "";
$user = "user$n";
$coll = "coll$n";

$mysqli = getWolkDatabase();
/*
// create account
$cmd = "$bin createaccount --httpport=$port --server=c0.wolk.com --waitfortx $user";
echo "$cmd\n";
$a = exec($cmd);
echo "$a\n";

// mkdir
$cmd = "$bin mkdir --httpport=$port --server=c0.wolk.com --waitfortx wolk://$user/$coll";
echo "$cmd\n";
$a = exec($cmd);
echo "$a\n";
*/

for ($cnt = 1; $cnt< 10000; $cnt++) {
    // set
    $cmd = "$bin set --httpport=$port --server=c0.wolk.com wolk://$user/$coll/ticker $cnt";
    echo "$cmd\n";
    $a = exec($cmd);
    echo "SET($cnt): $a\n";
    $setDT = date("Y-m-d H:i:s", time());
    
    // get
    $done = false;
    $tries = 0;
    while ( ! $done ) {
      $cmd = "$bin get --httpport=$port --server=c0.wolk.com wolk://$user/$coll/ticker";
      echo "$cmd\n";
      $b = exec($cmd);
      echo "GET (Try $tries): $b (expecting $cnt)\n";
      $getDT = date("Y-m-d H:i:s", time());
      if ( intval($b) == $cnt ) {
         $done = true;
	 $success = 1;
      } else {
      	 $success = 0;
      }
      $sql = "insert into preemptivecheck ( tries, observed, expected, getDT, setDT, success ) values ('$tries', '$b', '$cnt', '$getDT', '$setDT', '$success')";
      if ( $mysqli->query($sql) ) {
         echo "recorded $sql\n";
      } else {
         echo "$sql\n".$mysqli->error."\n";;
      }
      $tries++;
    }
}
?>

