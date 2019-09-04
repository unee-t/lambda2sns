#!/bin/bash

STAGE=dev

show_help() {
cat << EOF
Usage: ${0##*/} [-p]

By default, deploy to dev environment on AWS account 812644853088

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

echo Connecting to ${STAGE^^}

MYSQL_PASSWORD=$(aws --profile $AWS_PROFILE ssm get-parameters --names UNTEDB_ROOT_PASS --with-decryption --query Parameters[0].Value --output text)
MYSQL_USER=$(aws --profile $AWS_PROFILE ssm get-parameters --names UNTEDB_ROOT_USER --with-decryption --query Parameters[0].Value --output text)
MYSQL_HOST=$(aws --profile $AWS_PROFILE ssm get-parameters --names UNTEDB_HOST --with-decryption --query Parameters[0].Value --output text)

echo mysql -s -h $MYSQL_HOST -P 3306 -u $MYSQL_USER --password=$MYSQL_PASSWORD
