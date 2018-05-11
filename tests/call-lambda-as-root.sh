#!/bin/bash

acc() {
	case $1 in
		dev)  echo 812644853088
		;;
		demo) echo 915001051872
		;;
		prod) echo 192458993663
		;;
	esac
}

domain() {
	case $1 in
		prod) echo auroradb.unee-t.com
		;;
		*) echo auroradb.$1.unee-t.com
		;;
	esac
}


for STAGE in demo dev prod
do

ssm() {
	aws --profile uneet-$STAGE ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}

echo mysql -h $(domain $STAGE) -P 3306 -u root --password=$(ssm MYSQL_ROOT_PASSWORD)

echo "CALL mysql.lambda_async( 'arn:aws:lambda:ap-southeast-1:$(acc $STAGE):function:alambda_simple',  '{ \"operation\" : \"$STAGE $(date)\" }' );"  |
	mysql -h $(domain $STAGE) -P 3306 -u root --password=$(ssm MYSQL_ROOT_PASSWORD)

done
