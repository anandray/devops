<?php
    require_once "Mail.php";

    $from = "adops@mdotm.com";
    $to = "anand@mdotm.com";
    $subject = "This is my subject";
    $body = "Hi,\n\nHow are you?";

;    $host = "localhost";
;    $username = "you@yourdomain.com";
;    $password = "yourpassword";
    $host = "smtp.gmail.com";
    $username = "adops@mdotm.com";
    $password = "M0r3L0v3!";

    $headers = array ('From' => $from,
      'To' => $to,
      'Subject' => $subject);    //FOR HTML FORMAT: $headers = array ('From' => $from, 'To' => $to, 'Subject' => $subject, 'Content-type' => 'text/html');
    $smtp = Mail::factory('smtp',
      array ('host' => $host,
        'auth' => true,
        'username' => $username,
        'password' => $password));

    $mail = $smtp->send($to, $headers, $body);

    if (PEAR::isError($mail)) {
      echo("<p>" . $mail->getMessage() . "</p>");
     } else {
      echo("<p>Message successfully sent!!</p>");
     }
?>
