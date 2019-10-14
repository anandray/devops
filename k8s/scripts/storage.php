<?php 
function getWolkDatabase($force = false)
{
	$mysqli = new mysqli("db03", "root", "1wasb0rn2!", "wolk");
	if ($mysqli->connect_errno) {
	    echo "Failed to connect to MySQL: (" . $mysqli->connect_errno . ") " . $mysqli->connect_error;
	}
	return $mysqli;
}

function get_testnet($stage, $restart = false, $cluster = "z") {
   global $mysqli;
   if ( $restart ) {
     $ts = time();
     $sql = "update testnet set networkID = '$ts' where stage = '$stage' and cluster = '$cluster'";
     if ( $res = $mysqli->query($sql) ) {
       echo "$sql\n";
     } else {
       echo $mysqli->error()."$sql\n";
     }
   }
   $sql = "select networkID from testnet where stage = '$stage' and cluster = '$cluster'";
   if ( $res = $mysqli->query($sql) ) {
        while ( $a = $res->fetch_object() ) {
 		return $a;
	}
   } else {
	#echo $mysqli->error();
   }
   echo "testnetID not found for stage $stage [$sql]\n";
   exit(0);
}

function getCluster() {
  $cluster = getenv("CEPHCLUSTER");
  if ( $cluster == "" ) {
    $cluster = "k";
  }
  return $cluster;
}
?>
