<?php
include "storage.php";
error_reporting(E_WARNING);

getWolkDatabase(true);

function get_aws_arn() {
  $sql = "select projectID, region from project where cloudprovider ='aws' and active=1 and lb is null";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      $cmd = "aws elbv2 describe-target-groups --region ".$a->region." --query TargetGroups[*].TargetGroupArn --output text";
      $b = array();
      $output = exec($cmd, $b);
      $result = $b[0];
      $projectID = $a->projectID;
      $sql = "update project set lb = '".$result."' where projectID ='$projectID'";
      echo "$sql\n\n";
      if ( mysql_query($sql) ) {
      }
    }
  } else {
    echo mysql_error();
    exit(0);
  }
  echo "\n\n";
}

get_aws_arn();
$sql = "select projectID, lb, region, cloudprovider from project where active=1 and lb is not null";
echo "$sql\n";
if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      if ($a->cloudprovider == "aws") {
	$arn = $a->lb;
	$region = $a->region;
	$cmd = "aws elbv2 describe-target-health --target-group-arn $arn --region $region";
	$arn = $a->arn;
	$region = $a->region;
	echo "$cmd\n";
	$b = array();
	$output = exec($cmd, $b);
	$str = implode(" ", $b);
	$out = json_decode($str);
	foreach ($out->TargetHealthDescriptions as $o) {
	  $target = $o->Target;
	  $targetHealth = $o->TargetHealth;
	  
	  $id = $target->Id;
	  $state = $targetHealth->State;
	  $sql = "update servers set healthcheck = '$state', healthcheckDT = Now() where id = '$id'";
	  if ( mysql_query($sql) ) {
	    echo "$sql\n";
	  } else {
	    echo mysql_error()." $sql\n";
	  }
	}
      } else {
	$projectID = $a->projectID;
	$cmd = "gcloud compute --project $projectID backend-services get-health `gcloud compute backend-services list --project $projectID | grep -v NAME | awk '{print$1}'` --global --format json";
	$b = array();
	$output = exec($cmd, $b);
	$str = implode(" ", $b);
	$out = json_decode($str);
	foreach ($out as $m) {
	  if ( isset($m->status) ) {
	    foreach ($m->status->healthStatus as $hs) {
	      $ipaddress = $hs->ipAddress;
	      $healthState = strtolower($hs->healthState);
	      $sql = "update servers set healthcheckDT = Now(), healthcheck = '".$healthState."' where privateip = '$ipaddress'";
	      if ( mysql_query($sql) ) {
		echo "$sql\n";
	      } else {
		echo mysql_error().$sql."\n";
	      }
	    }
	  }
	}

	echo "---\n";
      }
    }
}

?>
