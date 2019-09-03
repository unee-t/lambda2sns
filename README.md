[![Build Status](https://travis-ci.org/unee-t/lambda2sqs.svg?branch=master)](https://travis-ci.org/unee-t/lambda2sqs)

<img src="https://media.dev.unee-t.com/2019-09-03/lambda2sqs.png" alt="Lambda2SQS">

lambda2sqs originally started life as a bridge for [Aurora CALL
mysql.lambda_async](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Integrating.Lambda.html)
payloads to SNS to be subscribed to. It has evolved to co-ordinate our
Microservice architecture via a Amazon Simple Queue Service (SQS).

# Architecture

This is codified in [template.yaml](template.yaml) in AWS CloudFormation.

And deployed as [SAM
application](https://ap-southeast-1.console.aws.amazon.com/lambda/home?region=ap-southeast-1#/applications/lambda2sqs)
in order to orchestrate and co-ordinate the queues.

# Monitor the lambda functions

1. https://ap-southeast-1.console.aws.amazon.com/lambda/home?region=ap-southeast-1#/functions/ut_lambda2sqs_push?tab=monitoring
2. https://ap-southeast-1.console.aws.amazon.com/lambda/home?region=ap-southeast-1#/functions/ut_lambda2sqs_process?tab=monitoring
