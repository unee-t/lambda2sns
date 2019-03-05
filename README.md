[![Build Status](https://travis-ci.org/unee-t/lambda2sns.svg?branch=master)](https://travis-ci.org/unee-t/lambda2sns)

	lambda: arn:aws:lambda:ap-southeast-1:812644853088:function:alambda_simple
	sns: arn:aws:sns:ap-southeast-1:812644853088:atest

Setup [Aurora to trigger a lambda event](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Integrating.Lambda.html)

	SELECT lambda_sync(
		'arn:aws:lambda:ap-southeast-1:812644853088:function:alambda_simple',
		'{"operation": "ping"}');

Create an email subscription on the SNS topic:

https://ap-southeast-1.console.aws.amazon.com/sns/v2/home?region=ap-southeast-1#/topics/arn:aws:sns:ap-southeast-1:812644853088:atest

Then you should get an email of the JSON payload.

# Deploy and test

	make

# Setup

## lambda role

Assuming project.demo.json is created with correct profile.

	apex -r ap-southeast-1 --env demo init

## sns topic

	[hendry@t480s alambda]$ aws --profile uneet-demo sns create-topic --name atest
	{
		"TopicArn": "arn:aws:sns:ap-southeast-1:915001051872:atest"
	}
	[hendry@t480s alambda]$ aws --profile uneet-prod sns create-topic --name atest
	{
		"TopicArn": "arn:aws:sns:ap-southeast-1:192458993663:atest"
	}

How to subscribe:

	aws --profile uneet-prod sns subscribe --protocol email --topic-arn arn:aws:sns:ap-southeast-1:192458993663:atest --notification-endpoint youremail@example.com

Don't forget to confirm the subscription.

# Wonderful world of time wasting permission errors

	Lambda API returned error: Missing Credentials: Cannot instantiate Lambda Client

Read: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Integrating.Lambda.html#AuroraMySQL.Integrating.LambdaAccess

* https://console.aws.amazon.com/iam/home?region=ap-southeast-1#/roles/Aurora_access_to_lambda?section=permissions

* AWSLambdaFullAccess should simply have the managed policy **AWSLambdaFullAccess** attached.
* Modify IAM role on the Cluster
* Ensure the cluster parameter group has the arn:aws:iam::\*:role/Aurora_access_to_lambda defined in **aws_default_lambda_role**!

<img src=https://s.natalian.org/2018-05-11/lambda-aurora.png>
<img src=https://s.natalian.org/2018-05-11/1526021466_2558x1406.png>

## Lambda API returned error: Missing Credentials: Cannot instantiate Lambda Client

You need to **IAM roles to this cluster** https://ap-southeast-1.console.aws.amazon.com/rds/home?region=ap-southeast-1#dbclusters:
