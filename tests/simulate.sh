#!/bin/sh
test -f $1 || exit

# validate json
jq -e < $1 || exit

jq --argfile file $1 '.Message = ($file | tojson)' hook.json | curl http://localhost:3000/api/db-change-message/process

aws --profile uneet-dev sns publish --topic-arn arn:aws:sns:ap-southeast-1:812644853088:atest --message "$(cat $1)"
