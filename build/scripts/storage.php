<?php 
  /* php 7 
function getWolkDatabase($force = false)
{
	$mysqli = new mysqli("db03", "root", "1wasb0rn2!", "wolk");
	if ($mysqli->connect_errno) {
	    echo "Failed to connect to MySQL: (" . $mysqli->connect_errno . ") " . $mysqli->connect_error;
	}
	return $mysqli;
}
  */
function getWolkDatabase($force = false)
{
  $hname = php_uname('n'); // gethostname();
  $srv = "db03"; // 35.188.53.100"; // db03
  $theDB = mysql_connect($srv, "root", "1wasb0rn2!", $force);
  mysql_select_db("wolk", $theDB);
}

?>
