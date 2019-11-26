#!/usr/bin/php
<?
include "storage.php";
include "www-servers.php";
include_once "communications.php";

$storage = new Storage;
$storage->getMdotmDatabase();

$today = date("F j, Y, g:i a T");
$mail_subject = "CODE PUSH CCDEX (" . $name . ") - $today";
$mail_body = "Code was pushed using pushccdex to www-ccdex/api.crosschannel.com at " . $today . " to " . $cluster . ".  Details below:\n\n";
$mail_body .= "Pushed by: " . $name . "\n";
$mail_body .= "QA by: " . $qa . "\n";
$mail_body .= "Reason: " . $reason . "\n\n";

$mail_return = crosschannel_mail("adops@crosschannel.com", "engineering@crosschannel.com", $mail_subject, $mail_body);
if ($mail_return) echo("Message successfully sent!!\n");
?>
