<?
$msgs = 0;
while ( $msgs < 60 ) {
  $j = new StdClass;
  $j->siteID = md5(date("s", time()));
  $j->adID = substr(md5(date("i", time())), 0, 8);
  $j->impressions = rand(1, 100);
  $j->earnings = rand(1, 1000)/1000;

  $str = json_encode($j);
  echo "$msgs: $str\n";
  openlog('cc-test', LOG_PID | LOG_ODELAY, LOG_LOCAL4);   
  syslog(LOG_INFO, $str);
  closelog();
  sleep(1);
  $msgs++;
}
?>
