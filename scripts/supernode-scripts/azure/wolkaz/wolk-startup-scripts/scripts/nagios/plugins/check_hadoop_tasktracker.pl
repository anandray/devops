#!/usr/bin/perl -w

use strict;
use Socket;
use FileHandle;


my $node;
my $warn_node;
my $crit_node;
my $http;
my $buf;
my ($host, $port, $url, $ip, $sockaddr);
my $ST_OK=0;
my $ST_WR=1;
my $ST_CR=2;
my $ST_UK=3;



$host = $ARGV[0];
$warn_node = $ARGV[1]; #WARNING when there is less than this number of nodes alive
$crit_node = $ARGV[2]; #CRITICAL when there is less than this number of nodes alive
#$port = 50070;
#$port = 50030;
#$url = '/machines.jsp?type=active';
$port = 50060;
$url = '/tasktracker.jsp';

# create socket
$ip = inet_aton($host) || die "CRITICAL - host($host) not found.\n";
$sockaddr = pack_sockaddr_in($port, $ip);
socket(SOCKET, PF_INET, SOCK_STREAM, 0) || die "CRITICAL - socket error.\n";

# connect socket
connect(SOCKET, $sockaddr) || die "CRITICAL - connect $host $port error.\n";
autoflush SOCKET (1);

print SOCKET "GET $url HTTP/1.0\n\n";

while ($buf=<SOCKET>) {
	$_ = $buf;
	# get the numobe of "Live Nodes" from the response of http request
	if( /<a href=\"machines.jsp\?type=active\">[0-9]+<\/a>/){
		$node = $&;
		$node =~ s/<[^>]*>//gs;
		$node =~ s/\s//g;
	}
}

close(SOCKET);

if ($node <= $crit_node ){
	print "CRITICAL - TaskTracker up and running: $node \n";
	exit($ST_CR);
} elsif ($node <= $warn_node ){
	print "WARNING - TaskTracker up and running: $node \n";
	exit($ST_WR);
}else{
	print "OK - TaskTracker up and running: $node \n";
	exit($ST_OK);
}

