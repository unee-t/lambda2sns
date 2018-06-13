#!/bin/sh
test -f $1 || exit

# Validate JSON event
jq -e < $1 || exit

# Test locally
jq --argfile file $1 '.Message = ($file | tojson)' hook.json | curl -X POST -d @- http://localhost:3000/api/db-change-message/process

# Use https://sh.unee-t.com/ to verify payload
#jq --argfile file $1 '.Message = ($file | tojson)' hook.json | curl -X POST -d @- https://sh.unee-t.com/hook

# Simulate SNS on dev
#aws --profile uneet-dev sns publish --topic-arn arn:aws:sns:ap-southeast-1:812644853088:atest --message "$(cat $1)"
