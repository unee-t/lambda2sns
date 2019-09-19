PROJECT = lambda2sqs
STACK_NAME ?= $(PROJECT)
AWS_REGION = ap-southeast-1
DEPLOY_S3_PREFIX = lambda2sqs

.PHONY: clean build test deploy

build: push-bin process-bin

push-bin: push/
	cd push; GO111MODULE=on go mod tidy; GOOS=linux GOARCH=amd64 go build -o ../push-bin .

process-bin: process/
	cd process; GO111MODULE=on go mod tidy; GOOS=linux GOARCH=amd64 go build -o ../process-bin .

push-logs:
	sam logs -n ut_lambda2sqs_push -t

process-logs:
	sam logs -n ut_lambda2sqs_process -t

invoke: push-bin
	sam local invoke Push -e tests/foo.json

destroy:
	aws cloudformation delete-stack \
		--stack-name $(STACK_NAME)

validate: template.yaml
	AWS_PROFILE=uneet-dev sam validate --template template.yaml

dev: build
	AWS_PROFILE=uneet-dev sam package --template-file template.yaml --s3-bucket dev-media-unee-t --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	AWS_PROFILE=uneet-dev sam deploy --template-file ./packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM

demo: build
	ls
	AWS_PROFILE=uneet-demo sam package --template-file template.yaml --s3-bucket demo-media-unee-t --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	AWS_PROFILE=uneet-demo sam deploy --template-file ./packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM --parameter-overrides DefaultSecurityGroup=sg-6f66d316 PrivateSubnets=subnet-0bdef9ce0d0e2f596,subnet-091e5c7d98cd80c0d,subnet-0fbf1eb8af1ca56e3

prod: build
	AWS_PROFILE=uneet-prod sam package --template-file template.yaml --s3-bucket prod-media-unee-t --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	AWS_PROFILE=uneet-prod sam deploy --template-file ./packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM --parameter-overrides DefaultSecurityGroup=sg-9f5b5ef8 PrivateSubnets=subnet-0df289b6d96447a84,subnet-0e41c71ad02ee7e99,subnet-01cb9ee064743ac56

lint:
	cfn-lint template.yaml

clean:
	rm -f push-bin process-bin
