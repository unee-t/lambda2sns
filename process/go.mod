module github.com/unee-t/lambda2sqs/process

go 1.12

require (
	github.com/apex/log v1.1.1
	github.com/aws/aws-lambda-go v1.13.0
	github.com/aws/aws-sdk-go-v2 v0.11.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/unee-t/env v0.0.0-20190513035325-a55bf10999d5
)

replace github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.7.0
