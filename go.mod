module github.com/unee-t/lambda2sns

go 1.12

require (
	github.com/apex/log v1.1.0
	github.com/aws/aws-lambda-go v1.10.0
	github.com/aws/aws-sdk-go-v2 v0.8.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/unee-t/env v0.0.0-20190513035325-a55bf10999d5
	google.golang.org/appengine v1.6.0 // indirect
)

replace github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.7.0
