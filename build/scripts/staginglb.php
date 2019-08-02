#!/usr/bin/php
<?php

include "storage.php";

error_reporting(E_ERROR);

function myexec($cmd, $run) {
  // echo "$cmd\n";
    if ( $run ) {
      $output = array();
      exec($cmd, $output);
      return ($output);
    }
}


getWolkDatabase(true);


$sql = "select projectID, region, node, instanceGroup, cloudprovider, lb from project where active=1 order by node";
if ( $res = mysql_query($sql) ) {
  while ( $a = mysql_fetch_object($res) ) {
    $projects[] = $a;
  }
} else {
  echo mysql_error();
  exit(0);
}


//$nstages = 6;
$nstages = 1;
$run = true;
for ($stage = 1; $stage <= $nstages; $stage++) {
  $PORT = 80 + $stage;
  foreach ($projects as $p) {
    $project = $p->projectID;
    $dc = str_replace("wolk-", "", $project); // us-central
    $instanceGroup = $p->instanceGroup;
    $region = $p->region;  // us-central-1
    $cloudprovider = $p->cloudprovider;
    $nodenumber = $p->node;
    if ( $cloudprovider == "gc" ) {
      // 1. CREATE/ASSOCIATE NAMED PORT $PORT WITH COMPUTE INSTANCE GROUP
      $cmd[] = "gcloud compute instance-groups set-named-ports --project $project --region=$region --named-ports=cloudstore-$PORT:$PORT wolk-$nodenumber-gc-$dc-datastore";
      // 2. CREATE TCP HEALTHCHECK
      $healthcheck = "cloudstore-$dc-healthcheck-$PORT";
      $cmd[] = "gcloud compute --project \"$project\" http-health-checks create \"$healthcheck\" --port \"$PORT\" --request-path \"/healthcheck\" --check-interval \"5\" --timeout \"5\" --unhealthy-threshold \"2\" --healthy-threshold \"2\"";

      // 3. CREATE BACKEND SERVICE
      $service = "cloudstore-$dc-gc-$PORT";
      $cmd[] = "gcloud compute --project $project backend-services create $service --global --http-health-checks $healthcheck --load-balancing-scheme=EXTERNAL --port-name=cloudstore-$PORT --protocol=HTTP";

      // 4. ADD BACKENDS TO BACKEND SERVICE
      $instancegroup = "wolk-$nodenumber-gc-$dc-datastore";
      $instancegroupregion = $region;
      $cmd[] = "gcloud compute backend-services add-backend $service --instance-group=$instancegroup --instance-group-region=$instancegroupregion --project $project --balancing-mode=UTILIZATION --global --max-utilization=0.8 --max-rate-per-instance=1000";

      // 5. Add URL MAP
      $cmd[] = "gcloud compute url-maps create $service --default-service $service --description \"Backend Service for LB on port $PORT\"  --project $project";

      // 6. CREATE HTTPS TARGET PROXY USING THE ABOVE URL MAP
      $httpsproxy = "cloudstore-$dc-gc-target-proxy-https-$PORT";
      $cmd[] = "gcloud compute --project=$project target-https-proxies create $httpsproxy --url-map=$service --ssl-certificates=wildcard-wolk-com";

      // 7. Create Static IP
      $staticip = "cloudstore-$dc-gc-global-ip-$PORT";
      $cmd[] = "gcloud beta compute --project=$project addresses create $staticip --global --network-tier=PREMIUM";

      // 8. Create GLOBAL forwarding-rules
      $rule = "cloudstore-$dc-https-$PORT";
      $cmd[] = "gcloud compute --project=$project forwarding-rules create $rule --global --address=$staticip --ip-protocol=TCP --ports=443 --target-https-proxy=$httpsproxy";
    } else if ( $cloudprovider == "aws" ) {
      $target_group_arn = $p->lb;
      // 1. Create Target Group
      $c = 'aws ec2 describe-vpcs --region='.$region.' | grep -i VpcId | cut -d"\"" -f4';
      $vpc_id = implode(" ", myexec($c, true));
      //echo "1: $c ===> $vpc_id\n";
      $cmd[] = "aws elbv2 --region $region create-target-group --name wolk-target-grp-$PORT-$region --protocol HTTP --port $PORT --vpc-id $vpc_id";

      // 2. Register Target group
      $c = 'aws autoscaling describe-auto-scaling-instances --region '.$region.' --query \'AutoScalingInstances[*].InstanceId\' --output text | awk -vORS=, \'{ print $1 }\' | sed \'s/,/\ /g\' | sed \'s/$/\n/\'';
      $instance_ids = implode(" ", myexec($c, true));
      //echo "2: $c ===> $instance_ids \n";
      $cmd[] = "aws elbv2 register-targets --target-group-arn $target_group_arn --targets Id=$instance_ids --region $region";

      // 3. Create Listener
      // $c = 'aws elbv2 describe-target-groups --region '.$region.' --query TargetGroups[*].TargetGroupArn  | grep -E -v "\-81\-|\-82\-|\-83|\-84\-|\-85\-" | grep aws';
      $c = 'aws elbv2 --region '.$region.' describe-load-balancers --query "LoadBalancers[].LoadBalancerArn" | grep -E -v "\-81\-|\-82\-|\-83|\-84\-|\-85\-" | grep aws';
      $loadbalancer_arn = implode(" ", myexec($c, true));
      //echo "3: $c ===> $loadbalancer_arn\n";
      $cmd[] = "aws elbv2 --region $region create-listener --load-balancer-arn $loadbalancer_arn --protocol HTTP --port $PORT --default-actions Type=forward,TargetGroupArn=$target_group_arn";
    } else {
      echo "UNKNOWN CLOUD PROVIDER $cloudprovider";
    }
  }
}
foreach ($cmd as $c) {
  //  myexec($cmd, $run);
  echo "$c\n";
}


?>
