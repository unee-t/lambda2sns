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

if test "$1"
then
	apex -r ap-southeast-1 --env prod logs -F "{ $.fields.error = \"$1\" }" --since $since
else
	apex -r ap-southeast-1 --env prod logs -F "{ $.level = \"error\" }" --since $since
fi
