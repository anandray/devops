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
   $sql = "select networkID from network where stage = '$stage'";
   if ( $res = $mysqli->query($sql) ) {
        while ( $a = $res->fetch_object() ) {
 		return $a;
	}
   } else {
	#echo $mysqli->error();
   }
   echo "networkID not found for stage $stage [$sql]\n";
   exit(0);
}

function getCluster($i) {
   global $mysqli;
   $sql = "select cluster from servers where nodeNumber = '$i'";
   if ( $res = $mysqli->query($sql) ) {
        while ( $a = $res->fetch_object() ) {
 	   return $a->cluster;
	}
   } else {
	#echo $mysqli->error();
   }
   return $cluster;
}
?>
