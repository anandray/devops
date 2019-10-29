#!/usr/bin/php
<?

$i = isset($argv[1]) ? $argv[1] : "NOTAVAIL";
switch ($i) {
 case "log01":
   echo "CRITICAL mopub is down";
   exit(2);
   break;
 case "log02":
   echo "OK log02 is OK";
   exit(0);
   break;
 default:
   echo "DEFAULT response with white snow -- [$i]";
   exit(2);
   break;
 }
echo "NO WAY [$i]";
exit(0);
?> 
