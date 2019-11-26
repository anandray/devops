<?php
include "storage.php";

error_reporting(E_ERROR);

getWolkDatabase(true);

for ($i=1; $i<7; $i++) {
  $logts = time()-86400*$i;
  $logDT = date("Y-m-d", $logts);
  $fn = "wolk-$logDT.csv";
  $datadir = "/var/log/wolk_billing/";
  $fullfn = "$datadir/$fn";
  $cmd = "gsutil cp gs://wolk_billing/$fn $fullfn";
  exec($cmd, $a);

  if ( file_exists($fullfn) ) {
    echo "Process: $fullfn\n";
    $cnt = 0;
    $num = 1;
    if ( $f = fopen($fullfn, "r") ) {
      $totalcost = 0;
      while ( ! feof($f) ) {
	$sa = fgetcsv($f);
	if ( $cnt == 0 ) {
	  foreach ($sa as $n => $val) {
	    $cols[$val] = $n;
	  }
	  print_r($cols);
	  $cnt++;
	} else {
	
	  $cnt++;
	  $cost = $sa[$cols["Cost"]];
	  $projectID = $sa[$cols["Project ID"]];
	  $description = $sa[$cols["Description"]];
	  //create table google_wolk_billing ( projectID varchar(32), logDT date, line_number int, description varchar(128), cost float default 0, primary key (projectID, line_number, logDT) );
	  if ( abs($credit) > 0 ) {
	    $cost += $credit;
	  }
	  $num++;
	  if ( abs($cost) > 0.01 ) {
	    $sql = "insert google_wolk_billing ( projectID, line_number, logDT, description, cost ) values ('$projectID', '$num', '$logDT', '$description', '$cost') on duplicate key update description = values(description)";
	    if ( mysql_query($sql) ) {
	      echo "$sql\n";
	    } else {
	      echo mysql_error();
	    }
	    $totalcost += $cost;
	  }
	}
      }
    }
  }
}
