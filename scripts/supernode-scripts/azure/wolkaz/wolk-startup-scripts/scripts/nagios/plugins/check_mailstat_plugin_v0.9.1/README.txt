check_mailstat.pl Copyright(C) Curu Wong 2010

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.


Description:

mailgraph is a good tool at analysing mail server log, it supports many 
MTA(including sendmail, postfix, exim,etc.), with support for mailscanner,spamassassin,etc.

As a whole, mailgraph can tell mail server statistics like how many messagessent/ received/ bounced/ rejected/ virus/ spam per time interval, which can 
be used as a sign of mail server healthy. 

This plugin includes a patch for mailgraph so that it will also output its statisticscounter to an external file(plus the rra file),and a check_mailstat.pl which check 
the stat counter to see if it's ok, emit WARN/CRITICAl result if not.
It can run on nagios server, or on remote server via NRPE.

Changelog:
Version 0.9.1 2011-03-17
 --add performance data output
 --add a sample PNP4Nagios template extra/check_mailstat.php
 
Version 0.9 2010-12-18

Install:

1. Install mailgraph on the server to be monitored, then patch it with mailgraph.patch
patch -b <path_to_mailgraph> mailgraph.patch
Many distribution included mailgraph is located at /usr/sbin/mailgraph. so you can:
patch -b /usr/sbin/mailgraph mailgraph.patchthen restart mailgraph.send out a test email, wait until a file appears at /var/tmp/mailstat. it will take
some time if your mail log is too large.

2. Also on host to be monitored,copy chech_mailstat.pl to your plugin directory and 
use it(you know how to use anagios plugin,right? ).
run check_mailstat.pl -h to get command usage.you need to use NRPE to execute this check if Nagios server is not on the same system as
mail server.

3. I recommend that in your service definition, set max_check_attempts to 1(or max_attempts 
in Nagios v2) to prevent retry this service check. But this is not required, sometimes, 
retry is OK, when some one is flushing your email server with bunches of message, you need
to retry to see if it's a tempeoral burst or continuous problem.

4. File extra/check_mailstat.php is a PNP4Nagios template, if you want to use this plugin
with PNP, you may need to copy that file to pnp4nagios/share/templates. 
