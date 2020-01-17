#!/bin/bash

ssm() {
	aws --profile ins-dev ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}

echo "select * from ut_map_external_source_users;" |
mysql -h auroradb.dev.unee-t.com -P 3306 -D unee_t_enterprise -u lambda_invoker --password=$(ssm LAMBDA_INVOKER_PASSWORD)
