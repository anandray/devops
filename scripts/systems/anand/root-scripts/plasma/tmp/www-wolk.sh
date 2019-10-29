echo "
#"api-colorfulnotion-com-bp6t" host definition
 define host{
 	 host_name 	 	 	 api-colorfulnotion-com-bp6t 
 	 alias 	 	 	 	 API-COLORFULNOTION-COM
 	 address 	 	 	 104.154.67.17 
 	 contact_groups 	 	 oncall-admins,oncall-admins2
 	 check_command 	 	 	 check-host-alive
 	 max_check_attempts 	 	 10
 	 notification_interval 	 	 120
 	 notification_period 	 	 24x7
 	 notification_options 	 	 d,u,r
 }" > /usr/local/nagios/etc/objects/servers/api-colorfulnotion-com-bp6t.cfg
echo "
#"api-colorfulnotion-com-q5rx" host definition
 define host{
 	 host_name 	 	 	 api-colorfulnotion-com-q5rx 
 	 alias 	 	 	 	 API-COLORFULNOTION-COM
 	 address 	 	 	 146.148.33.41 
 	 contact_groups 	 	 oncall-admins,oncall-admins2
 	 check_command 	 	 	 check-host-alive
 	 max_check_attempts 	 	 10
 	 notification_interval 	 	 120
 	 notification_period 	 	 24x7
 	 notification_options 	 	 d,u,r
 }" > /usr/local/nagios/etc/objects/servers/api-colorfulnotion-com-q5rx.cfg
echo "
#"api-colorfulnotion-com-r5vg" host definition
 define host{
 	 host_name 	 	 	 api-colorfulnotion-com-r5vg 
 	 alias 	 	 	 	 API-COLORFULNOTION-COM
 	 address 	 	 	 35.188.206.153 
 	 contact_groups 	 	 oncall-admins,oncall-admins2
 	 check_command 	 	 	 check-host-alive
 	 max_check_attempts 	 	 10
 	 notification_interval 	 	 120
 	 notification_period 	 	 24x7
 	 notification_options 	 	 d,u,r
 }" > /usr/local/nagios/etc/objects/servers/api-colorfulnotion-com-r5vg.cfg
