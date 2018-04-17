package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go/aws"
)

func handler(ctx context.Context, evt json.RawMessage) (string, error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}

	client := sns.New(cfg)
	req := client.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("Ctx: %s\nEvt: %s\n", ctx, evt)),
		TopicArn: aws.String("arn:aws:sns:ap-southeast-1:812644853088:atest"),
	})

	resp, err := req.Send()
	if err != nil {
		return "", err
	}

	// return fmt.Sprintf("Ctx: %s\nEvt: %s\n", ctx, evt), nil
	return fmt.Sprintf("Response: %s", resp), nil
}

func main() {
	lambda.Start(handler)
}
