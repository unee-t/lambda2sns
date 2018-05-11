package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func handler(ctx context.Context, evt json.RawMessage) (string, error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}

	stssvc := sts.New(cfg)
	input := &sts.GetCallerIdentityInput{}

	req := stssvc.GetCallerIdentityRequest(input)
	result, err := req.Send()
	if err != nil {
		return "", err
	}

	snssvc := sns.New(cfg)
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("Ctx: %s\nEvt: %s\n", ctx, evt)),
		TopicArn: aws.String(fmt.Sprintf("arn:aws:sns:ap-southeast-1:%s:atest", aws.StringValue(result.Account))),
	})

	resp, err := snsreq.Send()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Response: %s", resp), nil
}

func main() {
	lambda.Start(handler)
}
