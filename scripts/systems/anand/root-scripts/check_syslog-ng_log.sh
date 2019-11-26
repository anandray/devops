ssh log6b hostname && ssh log6b ps aux | grep syslog-ng | grep -v grep | grep supervising
ssh log00 hostname && ssh log00 ps aux | grep syslog-ng | grep -v grep | grep supervising
ssh log6 hostname && ssh log6 ps aux | grep syslog-ng | grep -v grep | grep supervising
ssh log6c hostname && ssh log6c ps aux | grep syslog-ng | grep -v grep | grep supervising
