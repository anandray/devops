#!/usr/bin/perl -w

# $Id
#------------------------------------------------------------------------------#
# Check_multiaddr                                                              #
#                                                                              #
# Copyright Florent Vuillemin 2005 - lafumah@users.sourceforge.net             #
# Nagios and the Nagios logo are trademarks of Ethan Galsta                    #
#                                                                              #
# This 'metaplugin' for Nagios will execute a given program several times      #
# using different addresses (given as an argument) and send back a unique      #
# result based on the information gathered                                     #
#                                                                              #
#------------------------------------------------------------------------------#
#                                                                              #
# This program is free software; you can redistribute it and/or modify         #
# it under the terms of the GNU General Public License as published by         #
# the Free Software Foundation; either version 2 of the License, or            #
# (at your option) any later version.                                          #
#                                                                              #
# This program is distributed in the hope that it will be useful,              #
# but WITHOUT ANY WARRANTY; without even the implied warranty of               #
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                #
# GNU General Public License for more details.                                 #
#                                                                              #
# You should have received a copy of the GNU General Public License            #
# along with this program; if not, write to the Free Software                  #
# Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA   #
#------------------------------------------------------------------------------#

use strict;

my $TIMEOUT = 9;	# You might need to edit this parameter

my %STATE = ('OK' => 0, 'WARNING' => 1, 'CRITICAL' => 2, 'UNKNOWN' => 3);
my %PRIO = (0=>3, 1=>2, 2=>0, 3=>1);	# See comments 'PRIORITY' below


#------------------------------------------------------------------------------#
# Main Program                                                                 #
#------------------------------------------------------------------------------#

# Let's start with several checks :
# At least 1 argument provided ?
if (scalar @ARGV == 0) {
	short_usage();
	exit $STATE{'UNKNOWN'};
}

# Do you need help ?
if ($ARGV[0] eq "--help") {
	long_usage();
	exit $STATE{'OK'};
}
if ($ARGV[0] eq "-h") {
	short_usage();
	exit $STATE{'OK'};
}	

# First argument can be executed ?
if (! -x $ARGV[0]) {
	print "$0: ".$ARGV[0]." cannot be executed.\n";
	exit $STATE{'UNKNOWN'};
}

# Now we look for an argument which would be a set of adresses
my @addresses;
my $addr_pos = 0;

foreach (@ARGV) {
	if ($_ =~ /((\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}.?)+)/) {
		@addresses = split(/,/, $1);
		last;
	}
	$addr_pos++;	# Position of this argument
}

# No address set could be found
if (scalar @addresses == 0) {
	print "$0: Unable to find an address set, such as: 192.168.0.1,192.168.0.2\n";
	exit $STATE{'UNKNOWN'};
}

# Set up the communication pipe (used by the children processes to send the
# plugin outputs and states
pipe(READ,WRITE);

# Set up an timeout
local $SIG{ALRM} = sub {
	print "Timeout detected (".$TIMEOUT."s - you can edit its duration in $0).\n";
	exit $STATE{'UNKNOWN'}
};
alarm($TIMEOUT);

# Now we fork() as many children as we need
spawnchildren();

#------------------------------------------------------------------------------#
# This part will be executed by the father process only.                       #
# It receives all the results returned by the plugins and processes them to    #
# send back Nagios the service state                                           #
#------------------------------------------------------------------------------#

# Pipe in read-only
close WRITE;

my ($data, $addr, $state, $output);

# $best_* are the default variables to return if no result is better
my ($best_addr, $best_state, $best_output) =
	("No Address", 3, "No result returned by the plugin !");

#
# PRIORITY:
# In this part I consider that each state has a priority defined as follows :
# OK (0) > WARNING (1) > UNKNOWN (3) > CRITICAL (2)
#
# As a consequence, if at least one plugin return 'OK', the state returne to
# Nagios will necessarily be OK. On the contrary, if this metaplugin returns
# 'CRITICAL' to Nagios, then it means every plugin executed (ie on each
# address of the checked host) has returned 'CRITICAL'.
# 

# The loop ends when no child process remains connected to the pipe any longer
while (defined($data = <READ>)) {
	if ($data =~ /^.+<>.+<>.+$/) {
		($addr, $state, $output) = split(/<>/, $data);
	} else {
		($addr, $state, $output) = ("?", $STATE{'UNKNOWN'}, "$0 did not receive valid data from ".$ARGV[0]);
	}
	($best_addr, $best_state, $best_output) = ($addr, $state, $output)
		if ($state <= $best_state);
}

# We have all the results, it would be a waste to timeout now...
alarm(0);

chomp($best_output);

# Returned to Nagios :
print "$best_addr: $best_output\n";
exit $best_state;

#------------------------------------------------------------------------------#
# SpawnChildren                                                                #
# Forks as many children as needed and gives them an address to use            #
#------------------------------------------------------------------------------#

sub spawnchildren {
	my ($addr, $pid);

	foreach $addr (@addresses) {
	    $pid = fork();

		if ($pid == -1) {
			print "$0: Unable to fork ! Please check your process count.";
			exit $STATE{'UNKNOWN'};
			
		} elsif ($pid == 0) {
			execcmd($addr);
			exit 0;	# should never be executed
		}
	}
}

#------------------------------------------------------------------------------#
# ExecCmd                                                                      #
# Modifies the command line provided in @ARGV to target only one address       #
# Send the results to the father process using a pipe                          #
#------------------------------------------------------------------------------#

sub execcmd {
	my $addr = shift;
	my @cmd;
	
	close READ;		# Stop reading input pipe
	select WRITE;	# Use WRITE instead of STDOUT
	$| = 1;

	# Replaces the address set by $addr in the command line
	for (my $i = 0; $i < scalar @ARGV; $i++){
		push(@cmd, (($i == $addr_pos) ? $addr : $ARGV[$i]));
	}

	my $result = `@cmd;echo \$?`;
	
	# I expect the plugin's return value (\d+) at the end of the string
	if (!($result =~ /^((.*\n*)+)\n(\d+)\n*$/)) {
		chomp($result);
		$result =~ s/\n+/;/g;	# Removes the line feeds
		print "$addr<>".$STATE{'UNKNOWN'}."<>$result\n";
		exit 1;

	} else {
		$result = $1;
		my $state = $3;
		chomp($result);
		$result =~ s/\n+/;/g;   # Removes the line feeds
		print "$addr<>$state<>$result\n";
		exit 0;
	}		
}



#------------------------------------------------------------------------------#
#                                                                              #
#		Usage                                                                  #
#		Short version & long version                                           #
#		This is really a verbose part                                          #
#                                                                              #
#------------------------------------------------------------------------------#

sub short_usage {
	print <<END
Check_multiaddr - Abstraction plugin for hosts using multiple interfaces
Usage: $0 /path/to/my/plugin [my plugin arguments]

Instead of using a single IP address, replace it by a set of addresses
separated by commas. Example: 192.168.0.1,192.168.0.11,192.168.0.21

This plugin uses an inner timeout of $TIMEOUT sec. You can edit it manually
inside this file : my \$TIMEOUT = [value];

>> Try '$0 --help | more' for a much verbose help !
END
}

sub long_usage {
	print <<END
Check_multiaddr - Abstraction plugin for hosts using multiple interfaces
Usage: $0 /path/to/my/plugin [my plugin arguments]

Instead of using a single IP address, replace it by a set of addresses
separated by commas. Example: 192.168.0.1,192.168.0.11,192.168.0.21

This plugin uses an inner timeout of $TIMEOUT sec. You can edit it manually
inside this file : my \$TIMEOUT = [value];

********************************* EXAMPLE ************************************

Suppose you have a server with 2 network interfaces (using these IP addresses:
192.168.0.1 & 192.168.0.11) and executing a DNS server. If the first interface
goes down, the second one is used and reciprocally. This plugin allows you to
check the service no matter which interface (or which address) is used, by
testing each available address with the check_dns plugin.

If at least one instance of the DNS plugin returns 'OK', then we assume the
service is up. Else, Check_multiaddr will return the best state given by the
instanced plugins, using the following priority :
              OK   >   WARNING   >  UNKNOWN   >   CRITICAL

Now, to test our DNS service using multiple IP addresses:

---- 1. We define a new check command

Add the following section in your 'check commands' definition file (in my
case, checkcommands.cfg):

define command{
	command_name check_multiple_dns
	command_line \$USER1\$/check_multiaddr.pl \$USER1\$/check_dns -H \$ARG1\$ -s \$HOSTADDRESS\$
}
		
---- 2. We define our host with several addresses

... ie instead of typing :
               address 192.168.0.1
... in our host definition (typically in hosts.cfg), we will use :
               address 192.168.0.1,192.168.0.11

---- 3. Then we create a service 'DNS'

This service will use 'check_multiple_dns' as check command :
define service{
    service_description     DNS
    host_name               Server4
    check_command           check_multiple_dns!www.google.com
	...
}
			
---- 4. Other services

Please note that Nagios will ALWAYS replace \$HOSTADDRESS\$ by THE TWO IP
addresses in check commands (including other services and host check). If
you also want to monitor each network interface, you can redefine a command
like this (it does not use \$HOSTADDRESS\$ but \$ARG1\$):

define command{
    command_name check_if_alive
    command_line \$USER1\$/check_ping -H \$ARG1\$ -w 3000.0,80% -c 5000.0,100%
}

And then define two more services on the host:
define service{
    service_description     eth0
    host_name               Server4
    check_command           check_if_alive!192.168.0.1   # 1st IP address
    ...
}
define service{
    service_description     eth1
    host_name               Server4
    check_command           check_if_alive!192.168.0.11  # 2nd IP address
    ...
}

---- Hope this plugin will be useful
     Florent Vuillemin - lafumah\@users.sourceforge.net

END
}

