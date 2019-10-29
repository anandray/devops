<?php
include "storage.php";

error_reporting(E_ERROR);
getWolkDatabase(true);

function get_alibaba_servers($project, $region, $nodenumber, $projecttable = "project")
{
  global $sqla;
  $cmd = "aliyuncli ecs DescribeInstances --RegionId $region --filter Instances.Instance[*] --output json";
  echo "$cmd\n";
  exec($cmd, $awsarr);
  $cloudprovider = "alibaba";
  $alibaba = json_decode(implode("", $awsarr));
  foreach ($alibaba as $j => $i) {
      $publicip = $i->PublicIpAddress->IpAddress[0];
      $instanceid = $i->InstanceId;
      $privateip = "";
      $hostname = $i->InstanceName;
      $datacenter = $region;
      echo "$publicip $hostname\n";
      $status = "RUNNING";
      $consensus = 0;
      if ( strstr($hostname, "-tablestore") ) {
	        $consensus = 1;
      } else {
	        $hostname = "$hostname-".$j; // wolk-ali-us-east-1-vm1 ==> wolk-ali-us-east-1-vm1-{0..3}
      }
      $datacenter = $region;
			$pushwolk = 1;
      $sql = "insert into servers ( hostname, project, id, datacenter, privateip, publicip,  status, createDT, lastUpdateDT, cloudprovider, consensus, nodenumber, pushwolk, projecttable) values (
		  '".$hostname."', '$project', '$instanceid', '".$datacenter."','".$privateip."','".$publicip."','$status', Now(), Now(), '$cloudprovider', '$consensus', '$nodenumber', '$pushwolk', '$projecttable' ) on duplicate key
		  update datacenter = values(datacenter), publicip = values(publicip), privateip = values(privateip), id = values(id), project = values(project), projecttable = values(projecttable),
			 lastUpdateDT = values(lastUpdateDT), status = values(status), cloudprovider = values(cloudprovider), pushwolk = values(pushwolk), consensus = values(consensus), nodenumber = values(nodenumber), hostname = values(hostname)";
      if ( strlen($publicip) > 0 ) {
      	echo "$sql\n";
      	$sqla[] = $sql;
      }

  }
  $cnt++;
}

function get_aws_servers($project, $region, $nodenumber, $projecttable = "project")
{
  global $sqla;
  $cmd = "aws ec2 describe-instances --region $region";
  echo "$cmd\n";
  exec($cmd, $awsarr);
  $cloudprovider = "aws";
  $aws = json_decode(implode("", $awsarr));
  foreach ($aws->Reservations as $r) {
    foreach ($r->Instances as $i) {
      $publicip = $i->PublicIpAddress;
      $instanceid = $i->InstanceId;
      $privateip = $i->PrivateIpAddress;
      $hostname = "$publicip";
      foreach ($i->Tags as $t) {
      	if ( $t->Key == "Name" ) {
      	  $hostname = $t->Value;
      	}
      }
      echo "$publicip $privateip $name\n";
      $status = "RUNNING";
      $consensus = 0;
      if ( strstr($hostname, "-dynamo") ) {
	        $consensus = 1;
      }
      $datacenter = $region;

      if ( strstr($hostname, "wolk-") || $nodenumber > 0 ) {
				$pushwolk = 1;
      } else {
				$pushwolk = 0;
      }
      $sql = "insert into servers ( hostname, project, id, datacenter, privateip, publicip,  status, createDT, lastUpdateDT, cloudprovider, consensus, nodenumber, pushwolk, projecttable) values (
		  '".$hostname."', '$project', '$instanceid', '".$datacenter."','".$privateip."','".$publicip."','$status', Now(), Now(), '$cloudprovider', '$consensus', '$nodenumber', '$pushwolk', '$projecttable' ) on duplicate key
		  update datacenter = values(datacenter), publicip = values(publicip), privateip = values(privateip), id = values(id), project = values(project),  projecttable = values(projecttable),
			 lastUpdateDT = values(lastUpdateDT), status = values(status), cloudprovider = values(cloudprovider), pushwolk = values(pushwolk), consensus = values(consensus), nodenumber = values(nodenumber), hostname = values(hostname)";
      if ( strlen($publicip) > 0 ) {
      	echo "$sql\n";
      	$sqla[] = $sql;
      }
    }
  }
  $cnt++;
}

function get_azure_servers_consensus($project, $region, $nodenumber, $projecttable = "project")
{
  global $sqla;
  $cmd = "az network public-ip list -g $project --query \"[]\" --output json";
  echo "$cmd\n";
  exec($cmd, $arr);
  $cloudprovider = "azure";
  $azure = json_decode(implode("", $arr));
  foreach ($azure as $j => $i) {
      $publicip = $i->ipAddress;
      $instanceid = $i->InstanceId;
      $privateip = "";
      $hostname = $i->name;
      $datacenter = $region;
      echo "$publicip $hostname\n";
      $status = "RUNNING";
      $consensus = 0;
      if ( $j == 0 ) {
	        $consensus = 1;
      		$datacenter = $region;
      		$pushwolk = 1;
      		$sql = "insert into servers ( hostname, project, id, datacenter, privateip, publicip,  status, createDT, lastUpdateDT, cloudprovider, consensus, nodenumber, pushwolk, projecttable) values (
      		  '".$hostname."', '$project', '$instanceid', '".$datacenter."','".$privateip."','".$publicip."','$status', Now(), Now(), '$cloudprovider', '$consensus', '$nodenumber', '$pushwolk',  '$projecttable') on duplicate key
      		  update datacenter = values(datacenter), publicip = values(publicip), privateip = values(privateip), id = values(id), project = values(project),  projecttable = values(projecttable),
      			 lastUpdateDT = values(lastUpdateDT), status = values(status), cloudprovider = values(cloudprovider), pushwolk = values(pushwolk), consensus = values(consensus), nodenumber = values(nodenumber), hostname = values(hostname)";
      		if ( strlen($publicip) > 0 ) {
      		  echo "$sql\n";
      		  $sqla[] = $sql;
      		}
      }
  }
  $cnt++;
}

function get_azure_servers_storage($project, $region, $nodenumber, $lb, $projecttable = "project")
{
  global $sqla;

  $cmd = "az vmss list-instances -g $project -n $lb -o json";
  echo "$cmd\n";
  exec($cmd, $arr);
  $cloudprovider = "azure";
  $azure = json_decode(implode("", $arr));
  foreach ($azure as $j => $i) {
    $names[$j] = $i->name;
  }

  $arr = array();
  $cmd = "az vmss list-instance-public-ips -g $project -n $lb -o json";
  echo "$cmd\n";
  exec($cmd, $arr);
  $cloudprovider = "azure";
  $azure = json_decode(implode("", $arr));
  foreach ($azure as $j => $i) {
      $publicip = $i->ipAddress;
      $instanceid = $i->resourceGuid;
      $privateip = "";
      $hostname = $names[$j];
      $datacenter = $region;
      echo "$publicip $hostname\n";
      $status = "RUNNING";
      $consensus = 0;
      $datacenter = $region;
      $pushwolk = 1;
      $sql = "insert into servers ( hostname, project, id, datacenter, privateip, publicip,  status, createDT, lastUpdateDT, cloudprovider, consensus, nodenumber, pushwolk, projecttable) values (
		  '".$hostname."', '$project', '$instanceid', '".$datacenter."','".$privateip."','".$publicip."','$status', Now(), Now(), '$cloudprovider', '$consensus', '$nodenumber', '$pushwolk' ) on duplicate key
		  update datacenter = values(datacenter), publicip = values(publicip), privateip = values(privateip), id = values(id), project = values(project),  projecttable = values(projecttable),
			 lastUpdateDT = values(lastUpdateDT), status = values(status), cloudprovider = values(cloudprovider), pushwolk = values(pushwolk), consensus = values(consensus), nodenumber = values(nodenumber), hostname = values(hostname)";
      if ( strlen($publicip) > 0 ) {
      	echo "$sql\n";
      	$sqla[] = $sql;
      }
  }
  $cnt++;
}

function get_nodenumber($hostname, $region)
{
  $sa = explode("-", $hostname);
  if ( count($sa) > 2 ) {
    return($sa[1]);
  }
  return -1;
}
function is_gc_consensus($hostname)
{
  $sa = explode("-", $hostname);
  $l = count($sa);
  $lp = $sa[$l-1];
  if ( strlen($lp) == 4 ) {
    echo "$hostname -- len:$l -- lp:$lp ---> FALSE\n";
    return(false);
  }
    echo "$hostname -- len:$l -- lp:$lp ---> TRUE\n";
  return(true);
}

function get_gc_servers($project, $nodenumberstart, $zones, $projecttable = "project")
{
	global $sqla;
	$cmd = "gcloud compute instances list  --project=$project --format=json";
	echo "$cmd\n";
	$arr = array();
	exec($cmd, $arr);
	$servers = json_decode(implode("", $arr));
  $cnt = 0;
  $cloudprovider = "gc";
	foreach ($servers as $s) {
			$hostname = $s->name;
      $consensus = ( is_gc_consensus($hostname) ) ? 1 : 0;
			$pushwolk = ( strstr($hostname, "wolk-") ) ? 1 : 0;
      $internalip = $s->networkInterfaces[0]->networkIP;
      $externalip = $s->networkInterfaces[0]->accessConfigs[0]->natIP;
      $ha = explode("-", $hostname);
      $nodenumber = $ha[1];
      if ( $nodenumber >= $nodenumberstart && ( $nodenumber <= $nodenumberstart + 8 ) ) {
        $status = "RUNNING";
  			$sql = "insert into servers ( hostname, project, datacenter, privateip, publicip,  status, createDT, lastUpdateDT, cloudprovider, consensus, nodenumber, pushwolk, projecttable)
               values (  '".$hostname."', '$project', '".$zones."', '$internalip', '$externalip',  '$status', Now(), Now(), '$cloudprovider', '$consensus', '$nodenumber', '$pushwolk' , '$projecttable')
                on duplicate key update datacenter = values(datacenter), project = values(project), projecttable = values(projecttable), publicip = values(publicip), privateip = values(privateip), lastUpdateDT = values(lastUpdateDT), status = values(status), cloudprovider = values(cloudprovider), consensus = values(consensus), nodenumber = values(nodenumber), pushwolk = values(pushwolk), hostname = values(hostname)";
  			$sqla[] = $sql;
  			$cnt++;
      }
	}
}

function get_registry_json($projecttable)
{
  $testnet = array();
  $sql = "SELECT hostname, dns, project, datacenter, publicip, cloudprovider, consensus, nodenumber, id
   FROM servers where projecttable = '$projecttable' order by consensus desc, nodenumber;";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      $a->consensus = ( $a->consensus > 0 );
      $a->nodenumber = intval($a->nodenumber);
      $testnet[] = $a;
    }
  } else {
    echo mysql_error();
    exit(0);
  }
  $str = "var registry=".json_encode($testnet)."\n";
  return $str;
}

function update_servers_dns($projecttable = "project1")
{
  // look up entries
  $sql = "select hostname, publicip, datacenter, cloudprovider, consensus, nodenumber, dns from servers where pushwolk=1 and projecttable = '$projecttable' order by nodenumber";
  echo "$dns\n";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      if ($a->consensus > 0 ) {
	$nm = "c".$a->nodenumber;
      } else {
	$nm = "s".$a->nodenumber."-".substr(md5($a->hostname), 0, 4);
      }
      $servers[$nm] = $a->publicip;
      $hostname[$nm] = $a->hostname;
      if ( $nm != $a->dns ) {
	echo "NM:[$nm]\tDNS:[".$a->dns."]\n";
	$sqla[] = "update servers set dns = '$nm' where hostname = '{$a->hostname}'";
      }
    }
  } else {
    echo mysql_error();
    exit(0);
  }
  foreach ($sqla as $sql) {
    if ( $res = mysql_query($sql) ) {
    }
  }

  /*  // delete all entries from the above
  $all = get_all_dns();
  foreach ($all as $name => $ip) {
    $nm = str_replace(".wolk.com", "", $name);
    if ( strlen($nm) > 1 ) {
      $fc = substr($nm, 0, 1);
      if ( $fc == "c" ) {
        $remaining = substr($nm, 1);
        if ( is_numeric($remaining) )  {
	  $a = get_dns($name);
	  if ( isset($servers[$nm]) ) {
	    echo "$nm\n";
	    foreach ($a as $rec) {
	      print_r($rec);
	      echo "delete_dns({$rec->id}) ".count($a)."\n";
	      delete_dns($rec->id);
	    }
	  }
        }
      }
    }
  }

  // create new entries
  foreach ($servers as $nm => $ip) {
    $res = create_dns("$nm.wolk.com", $ip);
  }
  */
}

function update_missing_lbs($projecttable = "project")
{
  $sql = "select projectID, node, cloudprovider, region, zones, instanceGroup from $projecttable  where active = 1 and ( lb is null or lb = '' )";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      $projectID = $a->projectID;
      $region = $a->region;
      $node = $a->node;
      $instanceGroup = $a->instanceGroup;
      $sql = false;
      if ( $a->cloudprovider == "alibaba" ) {
      	$cmd = "aliyuncli ess DescribeScalingGroups --RegionId $region --filter ScalingGroups.ScalingGroup[*].ScalingGroupId --output text";
      	$lb = exec($cmd);
      	if ( strlen($lb) > 0 ) {
      	  $sql = "update $projecttable set lb = '$lb' where projectID = '$projectID'";
      	}
      } else if ( $a->cloudprovider == "azure" ) {
      	$cmd = "az vmss list -g $projectID --query [*].name -o tsv";
      	$lb = exec($cmd);
      	if ( strlen($lb) > 0 ) {
      	  $sql = "update $projecttable set lb = '$lb' where projectID = '$projectID'";
      	}
      } else if ( $a->cloudprovider == "gc" ) {
        $cmd = "gcloud compute instances list --project=$projectID --format json";
        $arr = array();
        exec($cmd, $arr);
        $q = json_decode(implode("", $arr));
        if ( count($q) > 0 ) {
          $httpurlzonearr = explode("/", $q[0]->zone);
          $zones = array_pop($httpurlzonearr);
          $cmd = "gcloud compute instance-groups list --project=$projectID --format json";
          $arr = array();
          exec($cmd, $arr);
          $q = json_decode(implode("", $arr));
          if ( count($q) > 0 ) {
            foreach ($q as $q0) {
	      $lb = $q0->name;
	      $lbnarr = explode("-", $lb);
	      if ( count($lbnarr) > 1 && intval($lbnarr[1]) != $node ) {
		echo "skipping... $lb [does not match $node]\n";
	      } else {
		$sql = "update $projecttable set lb = '$lb', instanceGroup = '$lb', zones = '$zones' where projectID = '$projectID'";
		echo "$sql\n";
	      }
	    }
          }
        }
      }
      if ( $sql ) {
	      if ( mysql_query($sql) )  {
          echo "$sql SUCC\n";
        } else {
          echo "$sql FAIL\n";
            echo mysql_error();
        }
      }
    }
  } else {
    echo mysql_error();
    exit(0);
  }
}

// get the servers!
$projecttables = array(
		       "project" => "nodes.js",
		       "project1" => "nodes1.js",
		       // "project2" => "nodes2.js",
		       // "project3" => "nodes3.js",
		       // "project5" => "nodes5.js"
		       );
// update dns records [==> google DNS]

foreach ($projecttables as $projecttable => $nodefn) {
  update_servers_dns($projecttable);
}

foreach ($projecttables as $projecttable => $nodefn) {
    update_missing_lbs($projecttable);
}

foreach ($projecttables as $projecttable => $nodefn) {
  $sql = "select cloudprovider, projectID, region, node, lb, zones from $projecttable as project where cloudprovider = 'gc' and active = 1 order by node";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      if ( $a->cloudprovider == "aws" ) {
        get_aws_servers($a->projectID, $a->region, $a->node, $projecttable);
      } else if ( $a->cloudprovider == "azure" ) {
        get_azure_servers_consensus($a->projectID, $a->region, $a->node, $projecttable);
        get_azure_servers_storage($a->projectID, $a->region, $a->node, $a->lb, $projecttable);
      } else if ( $a->cloudprovider == "alibaba" ) {
        get_alibaba_servers($a->projectID, $a->region, $a->node, $projecttable);
      } else {
        get_gc_servers($a->projectID, $a->node, $a->zones, $projecttable);
      }
    }
  } else {
    echo mysql_error();
    exit(0);
  }
}

// remove old servers
$sqla[] = "delete from servers where lastUpdateDT < date_sub(Now(), interval 100 second)";
$success = true;
foreach ($sqla as $sql) {
  if ( mysql_query($sql) ) {
    echo $sql."\n";
  } else {
    $success = false;
    echo mysql_error()."$sql\n";
  }
}
// save registry json
$testnetdir = "/var/www/vhosts/cloudstore/testnet";
if ( file_exists($testnetdir) ) {
  foreach ($projecttables as $projecttable => $nodefn) {
    $fn = "$testnetdir/$nodefn";
    $str = get_registry_json($projecttable);
    echo "$str\n";
    file_put_contents($fn, $str);
  }
} else {
  echo "no dir $testnetdir";
}

?>
