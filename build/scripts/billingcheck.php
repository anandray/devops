<?php

if ( $f = fopen("t.csv", "r") ) {
  while ( $sa = fgetcsv($f) ) {
    $amt = trim($sa[2]);
    $desc = trim($sa[1]);
    if ( ( strstr($desc, "us-west1") || strstr($desc, "wolk-1307" ) ) )   {
      if ( $amt > 5 ) {
	echo "$desc\t$amt\n";
	$result[$desc] = $amt;
      }
      $sum += $amt;
    }
  }
}
arsort($result);
print_r($result);
echo "$sum\n";
?>