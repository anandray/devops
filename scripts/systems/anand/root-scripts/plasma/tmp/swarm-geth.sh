echo "
#"demo-swarm-wolk-com-80kh" host definition
 define host{
 	 host_name 	 	 	 demo-swarm-wolk-com-80kh 
 	 alias 	 	 	 	 SWARM
 	 address 	 	 	 demo-swarm-wolk-com-80kh 
 	 contact_groups 	 	 oncall-admins,oncall-admins2
 	 check_command 	 	 	 check-host-alive
 	 max_check_attempts 	 	 10
 	 notification_interval 	 	 120
 	 notification_period 	 	 24x7
 	 notification_options 	 	 d,u,r
 }" > /usr/local/nagios/etc/objects/servers/demo-swarm-wolk-com-80kh.cfg
