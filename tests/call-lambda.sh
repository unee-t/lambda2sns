#!/bin/bash

AWS_PROFILE=${AWS_PROFILE:-"ins-dev"}

if ! test -f "$1"
then
	echo Missing JSON payload
	exit 1
fi

json=$1

ssm() {
	aws --profile $AWS_PROFILE ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}

echo mysql -h $(ssm MYSQL_HOST) -P 3306 -u $(ssm LAMBDA_INVOKER_USERNAME) --password=$(ssm LAMBDA_INVOKER_PASSWORD)
if echo "CALL mysql.lambda_async( 'arn:aws:lambda:ap-southeast-1:$(ssm ACCOUNT_ID):function:ut_lambda2sqs_push', '$(jq -c . $json)' );" |
mysql -h $(ssm MYSQL_HOST) -P 3306 -u $(ssm LAMBDA_INVOKER_USERNAME) --password=$(ssm LAMBDA_INVOKER_PASSWORD)
then
	echo YES
else
	echo NO
fi
