#!/bin/bash

since="10m"
OPTIND=1

while getopts S: opt; do
    case $opt in
        S)  since=$OPTARG
            ;;
        *)
            exit 1
            ;;
    esac
done
shift "$((OPTIND-1))"

if test -z "$AWS_PROFILE"
then
	echo AWS_PROFILE unset
	exit 1
fi
echo Profile: $AWS_PROFILE

env=$(echo $AWS_PROFILE | cut -c 7-)

if test "$1"
then
	apex -r ap-southeast-1 --env $env logs -F "{ $.fields.error = \"$1\" }" --since $since
else
	apex -r ap-southeast-1 --env $env logs -F "{ $.level = \"error\" }" --since $since
fi
