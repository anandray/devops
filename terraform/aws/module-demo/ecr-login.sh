#!/bin/bash
REGION="us-west-1"
`aws ecr get-login --no-include-email --region ${REGION}`
