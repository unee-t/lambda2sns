
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
	sam validate --template template.yaml

dev: build
	sam package --template-file template.yaml --profile $(AWS_PROFILE) --s3-bucket $(S3_BUCKET_NAME) --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	sam deploy --template-file ./packaged.yaml --profile $(AWS_PROFILE) --stack-name $(PROJECT) --capabilities CAPABILITY_IAM --parameter-overrides DefaultSecurityGroup=$(DEFAULT_SECURITY_GROUP) PrivateSubnets=$(PRIVATE_SUBNETS)

demo: build
	ls
	sam package --template-file template.yaml --profile $(AWS_PROFILE) --s3-bucket $(S3_BUCKET_NAME) --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	sam deploy --template-file ./packaged.yaml --profile $(AWS_PROFILE) --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM --parameter-overrides DefaultSecurityGroup=$(DEFAULT_SECURITY_GROUP) PrivateSubnets=$(PRIVATE_SUBNETS)

prod: build
	sam package --template-file template.yaml --profile $(AWS_PROFILE) --s3-bucket $(S3_BUCKET_NAME) --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	sam deploy --template-file ./packaged.yaml --profile $(AWS_PROFILE) --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM --parameter-overrides DefaultSecurityGroup=$(DEFAULT_SECURITY_GROUP) PrivateSubnets=$(PRIVATE_SUBNETS)

lint:
	cfn-lint template.yaml

clean:
	rm -f push-bin process-bin
