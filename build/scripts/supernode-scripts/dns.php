<?php
error_reporting(E_ERROR);

define(CLOUDFLARE_ZONEID, "176f94ebbe5570fc89d7063210eb0bca");
define(CLOUDFLARE_AUTHKEY, "3c094624db45767d658e9f3ff9f5f4e921117");
define(CLOUDFLARE_EMAIL, "cloud@wolk.com");

function get_all_dns()
{

  $done = false;
  $page = 1;
  while ( ! $done ) {
    $cmd = 'curl -s -X GET "https://api.cloudflare.com/client/v4/zones/'.CLOUDFLARE_ZONEID.'/dns_records?page='.$page.'&per_page=100&order=type&direction=asc&type=A" -H "Content-Type:application/json" -H "X-Auth-Key:'.CLOUDFLARE_AUTHKEY.'" -H "X-Auth-Email:'.CLOUDFLARE_EMAIL.'"';
    // echo "$cmd\n";
    $arr = array();
    exec($cmd, $arr);
    $res = json_decode(implode("", $arr));
    if ($res->success) {
      $cnt = count($res->result);
      foreach ( $res->result as $i => $r) {
        if ( $r->type == "A" ) {
          $a[$r->name][] = $r->content;
        }
      }
    }
    if ( $cnt < 100 ) {
      $done = true;
    } else {
      $page++;
    }
  }
  return($a);
}

function get_dns($name)
{
  $cmd = 'curl -q -s -X GET "https://api.cloudflare.com/client/v4/zones/'.CLOUDFLARE_ZONEID.'/dns_records?page=1&per_page=20&order=type&direction=asc&name='.$name.'&type=A" -H "Content-Type:application/json" -H "X-Auth-Key:'.CLOUDFLARE_AUTHKEY.'" -H "X-Auth-Email:'.CLOUDFLARE_EMAIL.'"';
  // echo "$cmd\n";
  exec($cmd, $arr);
  $res = json_decode(implode("", $arr));
  if ($res->success) {
    return($res->result);
  }
  return(false);
}

function create_dns($name, $ip)
{
  $cmd = 'curl -X POST "https://api.cloudflare.com/client/v4/zones/'.CLOUDFLARE_ZONEID.'/dns_records" -H "X-Auth-Email: cloud@wolk.com" -H "X-Auth-Key:'.CLOUDFLARE_AUTHKEY.'" -H "Content-Type: application/json" --data \'{"type":"A","name":"'.$name.'","content":"'.$ip.'","ttl":1}\'';
  // echo "$cmd\n";
  exec($cmd, $arr);
  $res = json_decode(implode("", $arr));

  if ( $res->success ){
    //echo "ID: ".$res->result->id."\n";
    return($res->result);
  }
  return(false);
}

function update_dns($id, $name, $ip)
{
  $cmd = 'curl -X PUT "https://api.cloudflare.com/client/v4/zones/'.CLOUDFLARE_ZONEID.'/dns_records/'.$id.'" -H "X-Auth-Email: cloud@wolk.com" -H "X-Auth-Key:'.CLOUDFLARE_AUTHKEY.'" -H "Content-Type: application/json" --data \'{"type":"A","name":"'.$name.'","content":"'.$ip.'","ttl":1}\'';
  // echo "$cmd\n";
  exec($cmd, $arr);
  $res = json_decode(implode("", $arr));
  if ( $res->success ){
    return($res->result);
  }
  return(false);
}

function delete_dns($id)
{
  $cmd = 'curl -X DELETE "https://api.cloudflare.com/client/v4/zones/'.CLOUDFLARE_ZONEID.'/dns_records/'.$id.'" -H "X-Auth-Email: '.CLOUDFLARE_EMAIL.'"      -H "X-Auth-Key: '.CLOUDFLARE_AUTHKEY.'"';
  echo "$cmd\n";
  exec($cmd, $arr);
  $res = json_decode(implode("", $arr));
  if ( $res->success ){
    return($res);
  }
  return(false);
}

function execute_deletes($deletes)
{
  foreach ($deletes as $nm) {
    $nm = trim($nm);
    echo "$nm\n";
    $a = get_dns($nm);
    echo($a->id)."\n";
    delete_dns($a->id);
  }
}
//print_r(get_all_dns());
//$res = create_dns("c2.wolk.com", "47.74.245.250");
//print_r($res);

//$res = get_dns("c3.wolk.com");
//print_r($res);

//$res = update_dns("e0032a197fe096eff7c82074fe5c466d", "c3.wolk.com", "13.76.45.239");
//print_r($res);

//$res = delete_dns("6eb5fb0e678943e0c7278c267c91cd50")
//print_r($res);

// execute_deletes(file("deletes"));

?>
