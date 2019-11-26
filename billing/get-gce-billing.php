<?php
include_once "storage.php";
error_reporting('E_WARN');
getWolkDatabase();
date_default_timezone_set('America/Los_Angeles');

for ($i=1; $i<7; $i++) {
  $logts = time()-86400*$i;
  $logDT = date("Y-m-d", $logts);
  $fn = "wlk-$logDT.json";
  $datadir = "/var/log/wolk_billing/";
  $fullfn = "$datadir/$fn";
  $cmd = "gsutil cp  gs://wolk_billing/$fn $fullfn";
  exec($cmd, $a);
  $num = 0;
  if ( file_exists($fullfn) ) {
    echo "Process: $fullfn\n";
    $arr = json_decode(file_get_contents($fullfn));
    foreach ($arr as $s) {
      $cnt++;
      $cost = $s->cost->amount;
      if ( isset($s->credits) ) {
	$credit = $s->credits[0]->amount;
	if ( abs($credit) > 0 ) {
	  $cost += $credit;
	}
      } else {
	$credit = 0;
      }
      $lineItem = $num;
      $projectID = $s->projectId;
      $description = $s->description;
      $num++;
      if ( $cost > 0.01 || ( abs($credit) > 0 ) ) {
	$sql = "insert into google_wolk_billing ( projectID, lineItem, lineNumber, logDT, description, cost ) values ('$projectID', '$lineItem', '$num', '$logDT', '$description', '$cost') on duplicate key update description = values(description)";
	if ( mysql_query($sql) ) {
	  echo "$sql\n";
	} else {
	  echo "$sql\n".mysql_error();
	}
	$totalcost += $cost;
      }
    }
  }
}

?>