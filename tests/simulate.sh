#!/bin/sh
test -f "$1" || exit
API_ACCESS_TOKEN=$(ssm uneet-dev API_ACCESS_TOKEN)

# Validate JSON event
jq -e < $1 || exit

# Test locally
curl -i -H "Content-Type: application/json" -H "Authorization: Bearer blablabla" -X POST -d @$1 http://localhost:3000/api/db-change-message/process?accessToken=$API_ACCESS_TOKEN

# Use https://sh.unee-t.com/ to verify payload
#jq --argfile file $1 '.Message = ($file | tojson)' hook.json | curl -H "Content-Type: text/plain" -X POST -d @- https://sh.unee-t.com/hook

# Simulate SNS on dev
#aws --profile uneet-dev sns publish --topic-arn arn:aws:sns:ap-southeast-1:812644853088:atest --message "$(cat $1)"
