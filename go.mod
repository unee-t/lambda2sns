module github.com/unee-t/lambda2sns

go 1.12

require (
	github.com/apex/log v1.1.0
	github.com/aws/aws-lambda-go v1.10.0
	github.com/aws/aws-sdk-go-v2 v0.8.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/moul/http2curl v1.0.0
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/unee-t/env v0.0.0-20190513035325-a55bf10999d5
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529 // indirect
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5 // indirect
	golang.org/x/sys v0.0.0-20190509141414-a5b02f93d862 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190511041617-99f201b6807e // indirect
	google.golang.org/appengine v1.5.0 // indirect
)

replace github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.7.0
