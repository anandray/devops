#!/usr/bin/perl -w

# Number of binary log files left (not counting the one in use)
$num = 2;

# Mysql Server, User and Password
# (make sure the user can connect and has the "super" privilege)
$host = "localhost";
$user = "r00t";
$password = "NQc2tFSX9";
$socket = ""; # Optional

# --- You don't need to edit below here ---

if ($socket!~//) {
$sock = ";mysql_socket=".$socket;
} else {
$sock = "";
}

use DBI;

$dsn = "DBI:mysql::".$host.$sock;
$dbh = DBI->connect($dsn, $user, $password);

if (!$dbh) {
print "\nERROR connecting database - " . $DBI::errstr . ".\n";
exit;
}

$cmd = "SHOW MASTER LOGS";
$sth = $dbh->prepare($cmd);
$sth->execute;
$res = $sth->fetch;
if(!$res){
print "Erro ao selecionar log binario";
} else {
$i = 0;
while ($$res[0]) {
$i++;
$value[$i] = $$res[0];
$res = $sth->fetch;
}
}
$sth->finish;

$val = $i-$num;
if ($value[$val]) {
$cmd2 = "PURGE MASTER LOGS TO '$value[$val]'";
$sth2 = $dbh->prepare($cmd2);
$sth2->execute;
$sth2->finish;
}

$dbh->disconnect;
