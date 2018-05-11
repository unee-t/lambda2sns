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

	[hendry@t480s alambda]$ apex deploy
	   • config unchanged          env= function=simple
	   • updating function         env= function=simple
	   • updated alias current     env= function=simple version=5
	   • function updated          env= function=simple name=alambda_simple version=5
	[hendry@t480s alambda]$ apex invoke simple < event.json
	"Response: {\n  MessageId: \"894a7693-2239-554f-b78a-25dc9de5a5d8\"\n}"

Using [apex](http://apex.run/) with AWS_PROFILE **uneet-dev** in **ap-southeast-1**

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

# Wonderful world of time wasting permission errors the drive me nuts

	Lambda API returned error: Missing Credentials: Cannot instantiate Lambda Client

Read: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Integrating.Lambda.html#AuroraMySQL.Integrating.LambdaAccess
