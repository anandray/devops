
variable "tier" { default = "db-f1-micro" }
variable "name" { default = "db-devops-012020" }
variable "db_region" { default = "us-west1" }
variable "disk_size" { default = "20" }
variable "database_version" { default = "MYSQL_5_7" }
variable "user_host" { default = "%" }
variable "user_name" { default = "admin" }
variable "user_password" { default = "123456" }
variable "replication_type"  { default = "SYNCHRONOUS" }
variable "activation_policy" { default = "always" }
