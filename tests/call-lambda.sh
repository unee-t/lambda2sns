#!/bin/bash

STAGE=dev

show_help() {
cat << EOF
Usage: ${0##*/} [-p]

By default, call the dev environment on AWS account 812644853088

	-p          PRODUCTION 192458993663
	-d          DEMO 915001051872

EOF
}

while getopts "pd" opt
do
	case $opt in
		p)
			echo "PRODUCTION" >&2
			STAGE=prod
			;;
		d)
			echo "DEMO" >&2
			STAGE=demo
			;;
		*)
			show_help >&2
			exit 1
			;;
	esac
done
AWS_PROFILE=uneet-$STAGE
shift "$((OPTIND-1))"   # Discard the options and sentinel --

if ! test -f "$1"
then
	echo Missing JSON payload
	exit 1
fi

json=$1

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

ssm() {
	aws --profile $AWS_PROFILE ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text
}

echo mysql -h $(ssm MYSQL_HOST) -P 3306 -u $(ssm LAMBDA_INVOKER_USERNAME) --password=$(ssm LAMBDA_INVOKER_PASSWORD)
if echo "CALL mysql.lambda_async( 'arn:aws:lambda:ap-southeast-1:$(acc $STAGE):function:ut_lambda2sqs_push', '$(jq -c . $json)' );" |
mysql -h $(ssm MYSQL_HOST) -P 3306 -u $(ssm LAMBDA_INVOKER_USERNAME) --password=$(ssm LAMBDA_INVOKER_PASSWORD)
then
	echo YES
else
	echo NO
fi
