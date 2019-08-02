<?php 
function getWolkDatabase($force = false)
{
	$mysqli = new mysqli("db03", "root", "1wasb0rn2!", "wolk");
	if ($mysqli->connect_errno) {
	    echo "Failed to connect to MySQL: (" . $mysqli->connect_errno . ") " . $mysqli->connect_error;
	}
	return $mysqli;
}
?>
