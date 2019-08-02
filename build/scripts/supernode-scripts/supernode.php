<?php
function check_toml($server, $provider, $nodeType = "consensus", $publicip = "")
{
  $binary_file = "/root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk";
  $toml_file = "/root/go/src/github.com/wolkdb/cloudstore/wolk.toml";
  $cmd = "timeout 10 ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no $server \"cat $toml_file\"";
  echo "$server [$publicip @cloudprovider=$provider,nodeType=$nodeType] wolk.toml field check: $cmd\n";
  exec($cmd, $arr);
  // fields required for all toml
  $flds = array("ConsensusIdx", "NodeType", "GenesisFile", "SSLCertFile", "SSLKeyFile", "Provider");

  // fields 
  $provider_fields["gc"] = array("GoogleDatastoreProject" => 0, "GoogleDatastoreCredentials" => 1);
  $provider_fields["aws"] = array("AmazonRegion" => 0, "AmazonCredentials" => 1);
  $provider_fields["azure"] = array("MicrosoftAzureAccountName" => 0, "MicrosoftAzureAccountKey" => 0);
  $provider_fields["alibaba"] = array("AlibabaAccessKeyId" => 0, "AlibabaAccessKeySecret" => 0, "AlibabaRegion" => 0, "AlibabaEndpointURL" => 0);

  foreach ($provider_fields[$provider] as $fld => $isfile) {
    $flds[] = $fld;
  }
  // loop through and make sure each field is present
  foreach ($flds as $i => $fld) {
    $found = false;
    foreach ($arr as $str) {
      if ( strstr($str, $fld) ) {
	$found = true;
	$val = trim(str_replace(array($fld, " = "), array("", ""), $str), " \"");
	$parsed[$fld] = $val;
      }
    }
    if ( ! $found ) {
      $missing[$fld] = true;
    }
  }
  if ( count($missing) > 0 ) {
    foreach ($missing as $fld => $c) {
      echo "$server TOMLCHECK FAIL: MISSING $fld in wolk.toml [see $cmd]\n";
    }
  } else {
    echo "$server TOMLCHECK: PASS wolk.toml All expected fields present\n";
  }
  if ( $parsed["NodeType"] != $nodeType ) {
    echo "$server TOMLCHECK: FAIL expected $nodeType, observed [".$parsed["NodeType"]."]\n";
  }

  $filechecks = array("SSLCertFile", "SSLKeyFile", "GenesisFile");
  foreach ($provider_fields[$provider] as $provider_field => $isFile) {
    if ( $isFile > 0 ) {
      $filechecks[] = $provider_field;
    }
  }
  // check for files
  foreach ($filechecks as $i => $fld) {
    if ( isset($parsed[$fld]) ) {
       $fn = $parsed[$fld];
       if ( ! check_file_exists($server, $fn) ) {
         echo "$server FILECHECK: FAIL $fld $fn [see ssh $server \"cat $fn\"]\n";
       } else {
         echo "$server FILECHECK: PASS $fld $fn present\n";
       }
    }
  }
  
  $fn = $binary_file;
  if ( ! check_file_exists($server, $fn) ) {
    echo "$server FILECHECK: FAIL $fld $fn [see ssh $server \"cat $fn\"]\n";
  } else {
    echo "$server FILECHECK: PASS $fld $fn present\n";
  }

}

// check presence of file
function check_file_exists($server, $fn, $ts=10) 
{
  $cmd = "timeout $ts ssh -q -o ConnectTimeout=2 -o BatchMode=yes -o StrictHostKeyChecking=no  $server \"wc -l $fn\"";
  //  echo "$cmd\n";
  $a = exec($cmd);

  $sa = explode(" ", $a);
  $cnt = $sa[0];
  if ( $cnt > 0 ) {
    // echo "cnt: [$cnt]\n";
    return(true);
  }
  return(false);
}

?>
