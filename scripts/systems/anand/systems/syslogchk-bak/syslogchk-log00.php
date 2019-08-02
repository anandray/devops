<?  

$s = rand (20,30);
sleep ($s);
$time = date("Y-m-d H:i:s", time());
echo "\ntime:$time\n";
$res = "";

// LOG = /disk1/log/log/$YEAR/$MONTH/$DAY/$HOUR$MIN.log
// DEBUG = /disk1/log/debug/$YEAR/$MONTH/$DAY/$HOUR$MIN.debug
$cmd = "stat /disk1/log/debug/`date +'%Y'`/`date +'%m'`/`date +'%d'`/`date +'%H%M'`.debug | grep Modify | grep \"`date +'%Y-%m-%d %T'`\"";
$return = "";
$output = array();
$return_status = -1;
$return = exec($cmd, $output, $return_status);

if ( ($return_status == 0) && (strstr( $return, $time) ) ){
  echo "\n$time\n$return\nDEBUG:syslog-ng is running\n\n";
}else{
  echo "\nDEBUG CHECK FAILED\n$time\n$return\n";
  restart_syslog ();
}

// BID = /disk2/log/bid/$YEAR/$MONTH/$DAY/$HOUR$MIN.bid
$cmd = "stat /disk2/log/bid/`date +'%Y'`/`date +'%m'`/`date +'%d'`/`date +'%H%M'`.bid | grep Modify | grep \"`date +'%Y-%m-%d %T'`\"";
$return = "";
$output = array();
$return_status = -1;
$return = exec($cmd, $output, $return_status);

if ( ($return_status == 0) && (strstr( $return, $time) ) ){
  echo "\n$time\n$return\nBID:syslog-ng is running\n\n";
}else{
  echo "\nBID CHECK FAILED\n$time\n$return\n";
  restart_syslog ();
}


function restart_syslog () {
	//        $res .= "STEP 9 - syslog-ng (2) fail\n";
	$syslog_ng_cmd_4 = "pkill -9 syslog-ng && /sbin/service syslog-ng restart > /var/log/syslogchk.log 2>&1";
//	$syslog_ng_cmd_4 = "pkill -9 syslog-ng && /sbin/service syslog-ng restart2 2>&1";
	
//	$syslog_ng_cmd_4 = "/sbin/service syslog-ng restart 2>&1";
//	$syslog_ng_cmd_4 = "echo syslog-ng NOT running...";
	$return = "";
	$output = array();
	$return_status = -1;
	$return = exec($syslog_ng_cmd_4, $output, $return_status);
	//        return new_server_check_fail($res);
	echo "\n-------------1----------------\n";
	echo "\n------------------------------\n";
	echo "return: $return";
	echo "\n------------------------------\n";
	print_r($output);
	echo "\n------------------------------\n";
	echo "CONVERSION return_status: $return_status";
	echo "\n------------------------------\n";
	
	if  ($return_status == 1)  {
		sleep(2);
		$return = exec($syslog_ng_cmd_4, $output, $return_status);
		echo "\n--------------2---------------\n";
		echo "\n------------------------------\n";
		echo "return: $return";
		echo "\n------------------------------\n";
		print_r($output);
		echo "\n------------------------------\n";
		echo "CONVERSION return_status: $return_status";
		echo "\n------------------------------\n";
		
	}  
}

/*
if ( ($return_status == 0) && ($output[0]>0) ){
	$res .= "STEP 9 - syslog-ng (2) ok  |  \n";
} else {
	$res .= "STEP 9 - syslog-ng (2) fail\n";
	return new_server_check_fail($res);
}
*/

/*
 echo "\n------------------------------\n";
 echo "return: $return";
 echo "\n------------------------------\n";
 print_r($output);
 echo "\n------------------------------\n";
 echo "return_status: $return_status";
 echo "\n------------------------------\n";
 exit(0);
 */

?>
