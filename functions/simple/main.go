package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/unee-t/env"
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

	log.Infof("JSON payload: %s", evt)
	snssvc := sns.New(cfg)
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("%s", evt)),
		TopicArn: aws.String(fmt.Sprintf("arn:aws:sns:ap-southeast-1:%s:atest", aws.StringValue(result.Account))),
	})

	resp, err := snsreq.Send()
	if err != nil {
		return "", err
	}

	err = post2Case(cfg, evt)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Response: %s", resp), nil
}

// For event notifications https://github.com/unee-t/lambda2sns/tree/master/tests/events
func post2Case(cfg aws.Config, evt json.RawMessage) (err error) {
	e, err := env.New(cfg)
	if err != nil {
		return err
	}
	casehost := fmt.Sprintf("https://%s", e.Udomain("case"))
	APIAccessToken := e.GetSecret("API_ACCESS_TOKEN")
	log.Infof("Posting to: %s, payload %s, with key %s", casehost, evt, APIAccessToken)

	url := casehost + "/api/db-change-message/process?accessToken=" + APIAccessToken
	req, err := http.NewRequest("POST", url, strings.NewReader(string(evt)))
	if err != nil {
		log.WithError(err).Error("constructing POST")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+APIAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("POST request")
		return err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read body")
		return err
	}
	if res.StatusCode == http.StatusOK {
		log.Infof("Response code %d, Body: %s", res.StatusCode, string(resBody))
	} else {
		log.Warnf("Response code %d, Body: %s", res.StatusCode, string(resBody))
	}
	return err
}

func main() {
	lambda.Start(handler)
}
