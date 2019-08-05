<?php 
function getWolkDatabase($force = false)
{
      $hname = php_uname('n'); // gethostname();
      $srv = "db03"; // 35.188.53.100"; // db03
      $theDB = mysql_connect($srv, "root", "1wasb0rn2!", $force);
      mysql_select_db("wolk", $theDB);
}


?>
