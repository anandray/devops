#!/usr/bin/php
<?php
$wcloudloc = "/root/go/src/github.com/wolkdb/wolkjs/";
$appdir = "/root/go/src/github.com/wolkdb/wolkjs/schema/schema.org";
function checkPendingTxCount() {
	global $appdir, $port, $wcloudloc;
	$hp = "--httpport=$port";
	$pendingTxCmd = $wcloudloc."wcloud pendingtxcount $hp";
	//$pendingTxCmd = "/Users/rodneywitcher/src/github.com/wolkdb/wolkjs/wcloud pendingtxcount $hp";
	$output = array();
	exec($pendingTxCmd, $output);
	if( count($output) > 0 ) {
		$count = $output[0];
		return $count;
	} else {
		return -1; //todo: verify this case
	}
}
function myexec($cmd, $run) {
	echo "$cmd\n";
	if ( $run ) {
		$output = array();
		exec($cmd, $output);
		if ( count($output) > 0 ) {
			print_r($output);
		}
	}
}
function exec_cmds($cmds, $run = false) {
	foreach ($cmds as $cmd) {
    echo "$cmd\n";
		if( $run ) {
		  myexec($cmd, $run);
		}
	}
}
function setup($owner = "wolkadmin")
{
  global $appdir, $port, $wcloudloc;
  $hp = "--httpport=$port";
  $cmds = array();
  $cmds[] = $wcloudloc."wcloud createaccount $hp --waitfortx $owner";
  //$cmds[] = $wcloudloc."wcloud setdefault $hp $owner";
  // $cmds[] = $wcloudloc."wcloud mkdir $hp --waitfortx wolk://$owner/js";
  // $cmds[] = $wcloudloc."wcloud put $hp $appdir/wolk.js wolk://$owner/js/wolk.js";
  return $cmds;
}
function uploadapps($owner = "wolkadmin", $app = "schemas", $files = null)
{
  global $appdir, $port, $wcloudloc;
  $hp = "--httpport=$port";
  $cmds = array();
  $cmds[] = $wcloudloc."wcloud mkdir $hp --waitfortx wolk://$owner/$app";
  foreach ($files as $i => $fn) {
    $cmds[] = $wcloudloc."wcloud put $hp $appdir/$fn wolk://$owner/$app/$fn";
  }
  return $cmds;
}
$owner = "wolkadmin";
$ports = array(443);
$apps = array("schemas"); // => array("home.html", "navbar.html", "txns.html", "block.html", "footer.html"));
foreach ($ports as $port) {
	$pendingtxcount = checkPendingTxCount();
	echo "pendingtxcount =".$pendingtxcount."\n";
	if( $pendingtxcount > 50 ) {
		echo "\nPending transaction count = ".$pendingtxcount." exceeds limit.  EXITING ...\n";
		exit;
	}
  $cmds = setup($owner);
  exec_cmds($cmds, true);
  /*foreach ($apps as $app => $files) {
    $cmds = uploadapps($owner, $app, $files);
    exec_cmds($cmds);
  }*/
  foreach ($apps as $app) {
    $files = array();
    exec("ls -1 $appdir", $files);
	//TODO: make it recursive
    $cmds = uploadapps($owner, $app, $files);
    exec_cmds($cmds, true);
  }
}
?>
