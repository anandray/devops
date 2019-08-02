#!/usr/bin/php
<?php
include "storage.php";
error_reporting(E_ERROR);

getWolkDatabase(true);

function get_cloudproviderregions() {
  $sql = "select cloudprovider, region from cloudproviderregion where status > 0";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      $regions[$a->cloudprovider][$a->region] = 1;
    }
  } else {
    echo mysql_error();
  }
  return($regions);
}

function get_defaultprojectforregion()
{
  $sql = "select region, defaultProject from cloudproviderregion where cloudprovider = 'gc' and status = 1 and defaultProject is not NULL order by region";
  if ( $res = mysql_query($sql) ) {
    while ( $a = mysql_fetch_object($res) ) {
      $defaultprojectforregion[$a->region] = $a->defaultProject;
    }
  }
  return $defaultprojectforregion;
}

function generate_project($n, $nodes)
{
  if ( $n <= 5 ) {
    $cloudproviders = array("gc" => 1);
  } else if ( $n <= 10 ) {
    $cloudproviders = array("gc" => 1, "azure" => 1);
  } else if ( $n < 15 ) {
    $cloudproviders = array("alibaba" => 1);
  } else if ( $n < 20 ) {
    $cloudproviders = array("aws" => 1);
  } else {
    $cloudproviders = array("gc" => 1, "aws" => 1,  "alibaba" => 1, "azure" => 1);
  }
  $cloudproviderregions = get_cloudproviderregions();
  $defaultprojectforregion = get_defaultprojectforregion();
  $projecttable = "project$n";
  $drop_table_sql = "DROP TABLE IF EXISTS $projecttable;";
  $create_table_sql = "CREATE TABLE `$projecttable` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );";
  $sqla[] = $drop_table_sql;
  if ( $n < 10 ) {
    $sqla[] = $create_table_sql;
    for ( $i = 0; $i < $nodes;   ) {
      $node = $i + 8*$n;
      $cloudprovider = array_rand($cloudproviders);
      $regions = $cloudproviderregions[$cloudprovider];
      $region = array_rand($regions);
      if ( ! isset($regionpicked[$region]) ) {
        if ( $cloudprovider == "gc" ) {
          $projectID = $defaultprojectforregion[$region];
        } else {
          $projectID = "wolk-$node-$cloudprovider-$region";
        }
        $sqla[] = "insert into $projecttable (projectID, node, cloudprovider, region) values ('$projectID', '$node', '$cloudprovider', '$region');";
        $i++;
        $regionpicked[$region] = true;
      }
    }
  }
  return( $sqla );

}

for ($project = 4; $project < 10; $project++) {
  $sqla = generate_project($project, 8);
  foreach ($sqla as $sql) {
    echo $sql."\n";
  }
  echo "\n\n";
}
?>
