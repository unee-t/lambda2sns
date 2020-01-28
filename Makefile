PROJECT = lambda2sqs
STACK_NAME ?= $(PROJECT)
AWS_REGION = ap-southeast-1
DEPLOY_S3_PREFIX = lambda2sqs
AWS_PROFILE = ins-dev

define ssm
$(shell aws --profile $(AWS_PROFILE) ssm get-parameters --names $1 --with-decryption --query Parameters[0].Value --output text)
endef

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
	sam validate --profile $(AWS_PROFILE) --template template.yaml

deploy: build
	sam package --profile $(AWS_PROFILE) --template-file template.yaml --s3-bucket $(call ssm,S3_BUCKET_NAME) --s3-prefix $(DEPLOY_S3_PREFIX) --output-template-file packaged.yaml
	sam deploy --profile $(AWS_PROFILE) --template-file ./packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_IAM \
	--parameter-overrides DefaultSecurityGroup=$(call ssm,DEFAULT_SECURITY_GROUP) PrivateSubnets=$(call ssm,PRIVATE_SUBNET_1),$(call ssm,PRIVATE_SUBNET_2),$(call ssm,PRIVATE_SUBNET_3)

lint:
	cfn-lint template.yaml

clean:
	rm -f push-bin process-bin
