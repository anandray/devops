export TENANT_ID=ec5dd7b3-0e80-414c-ba2a-6495e1f57384
export SUBSCRIPTION=461fccf8-9e23-432e-bf03-75ba4073f3c4
az account set --subscription $SUBSCRIPTION
source <(az completion bash)
alias ll='ls -lta'
