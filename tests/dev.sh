#!/bin/bash
ssm() {
	aws --profile uneet-dev ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}
echo "CALL mysql.lambda_async( 'arn:aws:lambda:ap-southeast-1:812644853088:function:alambda_simple',  '{ \"operation\" : \"FOOBAR\" }' );"  |
	mysql -h auroradb.dev.unee-t.com -P 3306 -u root --password=$(ssm MYSQL_ROOT_PASSWORD)
