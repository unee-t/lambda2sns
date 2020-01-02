#!/bin/bash

STAGE=dev

show_help() {
cat << EOF
Usage: ${0##*/} [-p]

By default, deploy to dev environment on AWS account 

	-p          PRODUCTION 
	-d          DEMO 

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
#This Parameter is in aws-env.STAGE file
#AWS_PROFILE=uneet-$STAGE
shift "$((OPTIND-1))"   # Discard the options and sentinel --

echo Connecting to ${STAGE^^}
source aws-env.$STAGE
echo $STAGE
echo "SELECT * FROM log_lambdas ORDER BY id DESC LIMIT 10\G;" |
mysql -s -h $MYSQL_HOST -P 3306 -u $MYSQL_USER --password=$MYSQL_PASSWORD $UNTE_DB_NAME
