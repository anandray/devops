<?php
include("/root/go/src/github.com/wolkdb/cloudstore/build/scripts/storage.php");

$token = "89ed4ec42358dce6396bf188b611eeb0f2551b66";
$mysqli = getWolkDatabase();

function get_sha_status($sha) 
{
  global $token;
  $url = "https://$token:x-oauth-basic@api.github.com/repos/wolkdb/cloudstore/statuses/$sha";
  echo "$url\n";
  $ch = curl_init();
  curl_setopt($ch, CURLOPT_URL, $url);
  curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($ch, CURLOPT_HTTPHEADER, array(
					     'User-Agent: Github Robot'
					     ));
  $raw = curl_exec($ch);
  $statuses = json_decode($raw);
  curl_close($ch);     
  $statuslist = array();
  foreach ($statuses as $status) {
    $statuslist[$status->updated_at] = $status->state;
  }
  if ( count($statuslist) > 0 ) {
  krsort($statuslist);
    return reset($statuslist);
  }
  return false;
}

function get_pull_status() 
{
  // get pull requests
  global $token, $mysqli;
  $url = "https://$token:x-oauth-basic@api.github.com/repos/wolkdb/cloudstore/pulls";
  echo "$url\n";
  
  $ch = curl_init();
  curl_setopt($ch, CURLOPT_URL, $url);
  curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($ch, CURLOPT_HTTPHEADER, array(
					     'User-Agent: Github Robot'
					     ));
  $pullsraw = curl_exec($ch);
  $pulls = json_decode($pullsraw);
  curl_close($ch);     
  foreach ($pulls as $pull) {
    $pullNumber = $pull->number;
    $state = $pull->state;
    $sha = $pull->head->sha;
    $status = get_sha_status($sha);
    
    echo "#{$pull->number} : {$pull->head->sha} {$pull->state} $status\n";
    $sql = "insert into pulls ( pullNumber, sha, state, status ) values ( '$pullNumber', '$sha', '$state', '$status' ) on duplicate key update sha = values(sha), state = values(state), status = values(status)";
    if ($res = $mysqli->query($sql)) {
      if ( $status != "success" ) {
	wolktest($sha);
      }
    } else {
      echo $mysqli->error();
      exit(0);
    }
  }
}

function wolktest($sha) 
{
  $cmds = array();
  $cmds[] = "cd /root/go/src/github.com/wolkdb/cloudstore/wolk; git checkout $sha; make testwolk";
  print_r($cmds);
  
}

get_pull_status();
?>