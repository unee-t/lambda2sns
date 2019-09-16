[![Build Status](https://travis-ci.org/unee-t/lambda2sqs.svg?branch=master)](https://travis-ci.org/unee-t/lambda2sqs)

# How to check logs? Or what's in the queue?

Visit <https://ap-southeast-1.console.aws.amazon.com/lambda/home?region=ap-southeast-1#/applications/lambda2sqs?tab=overview>

Messages that had some sort of validation failure or repeatedly failed will be in the **Dead letter queue**.

# Architecture

<img src="https://media.dev.unee-t.com/2019-09-03/lambda2sqs.png" alt="Lambda2SQS">

lambda2sqs originally started life as a bridge for [Aurora CALL
mysql.lambda_async](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Integrating.Lambda.html)
payloads to SNS to be subscribed to. It has evolved to co-ordinate our
Microservice architecture via a Amazon Simple Queue Service (SQS).

This is codified in [template.yaml](template.yaml) in AWS CloudFormation.

And deployed as [SAM
application](https://ap-southeast-1.console.aws.amazon.com/lambda/home?region=ap-southeast-1#/applications/lambda2sqs)
in order to orchestrate and co-ordinate the queues.
