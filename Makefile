PROJECT = $(shell basename $(CURDIR))
STACK_NAME ?= $(PROJECT)
AWS_REGION = ap-southeast-1
DEPLOY_S3_BUCKET = dev-media-unee-t
DEPLOY_S3_PREFIX = lambda2sns

.PHONY: deps clean build test deploy

# https://blog.deleu.dev/leveraging-aws-sqs-retry-mechanism-lambda/

deps:
	go mod tidy

build: deps
	GOOS=linux GOARCH=amd64 go build -o lambda2sns .

test:
	go test ./...

logs:
	sam logs -n alambda_simple -t

destroy:
	aws cloudformation delete-stack \
		--stack-name $(STACK_NAME)

deploy: build
	sam validate --template template.yaml
	sam package --template-file template.yaml --s3-bucket $(DEPLOY_S3_BUCKET) --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	sam deploy --template-file ./packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM

lint:
	cfn-lint template.yaml
