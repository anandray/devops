<?
// goservice.php: restart 'bt' process if process running or start if not

if (isset($argv[1])) {
	$goBinFile = $argv[1]; //   bt    zn
} else {
	echo "\nno argv to goservice.php exit [must have argv (bt, zn,...)]\n\n";
	exit(0);
}

$cmd = "ps aux | grep $goBinFile | awk '{print $2,$11}'";
echo "\n$cmd\n";
exec($cmd, $a);
// find the pid of "./bt"
$pid = 0;
foreach ($a as $b) {
    $sa = explode(" ", $b);
    if ( $sa[1] == "./$goBinFile" || $sa[1] == "/var/www/vhosts/sourabh/crosschannel.com/bidder/bin/$goBinFile" || $sa[1] == "/var/www/vhosts/crosschannel.com/bidder/bin/$goBinFile" ) {
        $pid = $sa[0];
        echo "\n$pid\n";
    }
}
if ( $pid > 0 ) {
	
	// after "git fetch upstream ..." the bin file is not X so we must call "chmod +x"
	if ( gethostname() == "www6001" ) {
		$dev = "/sourabh";
	} else {
		$dev = "";
	}
	$dir = "/var/www/vhosts$dev/crosschannel.com/bidder/bin";
	$cmd = "chmod +x $dir/$goBinFile";
	echo "\n$cmd\n";
	exec($cmd, $arr);
	print_r($arr);	
	
    // kill existing process
    $cmd = "kill -USR2 $pid";
    echo "\n$cmd\n";
    exec($cmd, $arr);
    print_r($arr);
} else {
    // start new process
    if ( gethostname() == "www6001" ) {
        $dev = "/sourabh";
    } else {
        $dev = "";
    }
    $dir = "/var/www/vhosts$dev/crosschannel.com/bidder/bin";
    $cmd = "chmod +x $dir/$goBinFile";
    echo "\n$cmd\n";
    exec($cmd, $arr);
    print_r($arr);
    $cmd = "$dir/$goBinFile";
    $outputfile = "/var/log/$goBinFile.log";
    $pidfile = "/var/log/$goBinFile.pid";
    $cmd2 = sprintf("%s > %s 2>&1 & echo $! >> %s", $cmd, $outputfile, $pidfile);
    echo "\n$cmd2\n";
    exec($cmd2);
}


?>