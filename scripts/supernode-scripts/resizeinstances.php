#!/usr/bin/php
<?php

include "storage.php";

error_reporting(E_ERROR);

function myexec($cmd, $run) {
    echo "$cmd\n";
    if ( $run ) {
      $output = array();
      exec($cmd, $output);
      return ($output);
    }
}

getWolkDatabase(true);

$size = isset($argv[1]) ? $argv[1] : 1;
$projecttable = isset($argv[2]) ? $argv[2] : "project";


$sql = "select projectID, region, node, instanceGroup, cloudprovider, lb, zones from $projecttable where active=1  order by node";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $projects[] = $a;
  }
} else {
  echo mysql_error();
  exit(0);
}


$run = true;
foreach ($projects as $p) {
  $project = $p->projectID;
  $instanceGroup = $p->instanceGroup;
  $region = $p->region;
  $zones = $p->zones;
  $cloudprovider = $p->cloudprovider;
  $nodenumber = $p->node;
  if ( $cloudprovider == "gc" ) {
    // resize!
    if ( strlen($instanceGroup) > 0 ) {
      $cmds[] = "gcloud beta compute instance-groups managed resize $instanceGroup --region $region --size=$size --project $project";
    } else {
      echo "MISSING lb for project $project\n";
    }
  } else if ( $cloudprovider == "azure") {
    $vmscaleset = $p->lb;
    $cmds[] = "az vmss scale -g $project -n $vmscaleset --new-capacity $size 2>&1 &";
  } else if ( $cloudprovider == "alibaba") {
    $lb = $p->lb;
    $cmds[] = "aliyuncli ess ModifyScalingGroup --ScalingGroupId $lb --RegionId $region --MaxSize $size --MinSize $size";
  } else if ( $cloudprovider == "aws" ) {
    // resize!
    $cmds[] = "aws autoscaling update-auto-scaling-group --auto-scaling-group-name $instanceGroup --region $region --min-size $size --max-size $size";
  } else {
    echo "UNKNOWN CLOUD PROVIDER $cloudprovider";
  }
}
foreach ($cmds as $cmd) {
  myexec($cmd, $run);
}

$sql = "select id, hostname from servers where hostname like '%dynamo%'";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $consensus[$a->id] = $a->hostname;
  }
}
$size_with_consensus = $size+1;
foreach ($projects as $p) {
  $project = $p->projectID;
  $instanceGroup = $p->instanceGroup;
  $region = $p->region;
  $cloudprovider = $p->cloudprovider;
  $nodenumber = $p->node;
  if ( $cloudprovider == "aws" ) {
    $instances = array();
    while ( count($instances) != $size_with_consensus ) {
      $cmd_instances = "aws ec2 describe-instances --query 'Reservations[*].Instances[*].[InstanceId]' --filters \"Name=instance-state-name,Values=running\" --region $region --output text";
      $instances = myexec($cmd_instances, $run);
      echo "have ".count($instances)." running, expecting $size_with_consensus...\n";
      if ( count($instances) == $size_with_consensus ) {
      	foreach ($instances as $i => $instanceID) {
      	  if ( isset($consensus[$instanceID]) ) {
      	    echo "Skipping tagging of ".$consensus[$instanceID]."...\n";
      	  } else {
      	    $instanceID = trim($instanceID, " \",");
      	    $suffix = substr(md5($instanceID), 0, 4);
      	    $value = str_replace("wolk-autoscale-", "wolk-$nodenumber-aws-", $instanceGroup)."-".$suffix; // "wolk-autoscale-eu-west-2
      	    $cmdtag = "aws ec2 create-tags --resources $instanceID --tag \"Key=Name,Value=$value\" --region $region";
      	    echo "$cmdtag\n";
      	    $output = myexec($cmdtag, $run);
      	  }
      	}
      }
      sleep(1);
    }

    // look up the instances
    $cmd_instanceid = "aws autoscaling describe-auto-scaling-instances --region $region --query 'AutoScalingInstances[*].InstanceId' | grep -v -E \"\[|\]\" | awk -vORS=, '{print\"Id=\"$1}' | sed 's/,/\ /g'";
    $outputinstances = myexec($cmd_instanceid, $run);
    $instance_id = implode(" ", $outputinstances);

    // look up the target_group_arn
    // $cmd_targetarn = "aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn --output text | grep -E -v \"\-81\-|\-82\-|\-83|\-84\-|\-85\-\" | grep aws";
    $cmd_targetarn = "aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn | grep -E -v \"\-81\/|\-82\/|\-83\/|\-84\/|\-85\/\" | grep -v -E \"\[|\]\" | awk -vORS=, '{print$1}' | sed 's/,/\ /g'";
    // $cmd_targetarn = "aws elbv2 describe-target-groups --region $region --query TargetGroups[*].TargetGroupArn --output text | awk '{print$1}'";
    $output = myexec($cmd_targetarn, $run);
    $target_group_arn = implode(" ", $output);

    // register targets
    $cmd = "aws elbv2 register-targets --target-group-arn $target_group_arn --targets $instance_id --region $region";
    $output = myexec($cmd, $run);
  }
}

?>
