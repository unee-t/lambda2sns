#!/bin/bash

ssm() {
	aws --profile uneet-dev ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}

mysql -h auroradb.dev.unee-t.com -P 3306 -D unee_t_enterprise -u lambda_invoker --password=$(ssm LAMBDA_INVOKER_PASSWORD)
