package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	qURL = os.Getenv("SQS_URL")
)

func main() {
	log.SetHandler(jsonhandler.Default)
	lambda.Start(handler)
}

func handler(ctx context.Context, evt json.RawMessage) error {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.WithError(err).Error("failed to load AWS config")
		return err
	}
	log.WithField("raw", string(evt)).Info("incoming")
	base64Decoding, err := digest(evt)
	if err != nil {
		log.WithError(err).Error("failed to decode payload")
		return err
	}

	// deduplicationID, groupID, err := id(evt)
	// if err != nil {
	// 	log.WithError(err).Error("unable to ID payload")
	// 	return err
	// }

	svc := sqs.New(cfg)
	req := svc.SendMessageRequest(&sqs.SendMessageInput{
		MessageBody: aws.String(string(base64Decoding)),
		// MessageDeduplicationId: aws.String(deduplicationID),
		// MessageGroupId:         aws.String(groupID),
		QueueUrl: &qURL,
	})
	_, err = req.Send(context.TODO())
	if err != nil {
		log.WithError(err).Error("failed to send")
		return err
	}
	log.WithField("payload", string(base64Decoding)).Info("enqueued")
	return nil
}

func id(evt json.RawMessage) (deduplicationId, groupId string, err error) {
	// Check if action type
	type actionType struct {
		MEFIRequestID string `json:"mefeAPIRequestId"`
	}
	var act actionType
	err = json.Unmarshal(evt, &act)
	if err != nil {
		log.WithError(err).Error("failed to parse as actionType")
	}
	if act.MEFIRequestID != "" {
		return act.MEFIRequestID, "actionType", nil
	}
	// Check if notification type
	type notificationType struct {
		NotificationID string `json:"notification_id"`
	}
	var notification notificationType
	err = json.Unmarshal(evt, &notification)
	if err != nil {
		log.WithError(err).Error("failed to parse as notificationType")
	}
	if notification.NotificationID != "" {
		return notification.NotificationID, "notificationType", nil
	}
	return "", "", fmt.Errorf("failed to id payload: %+v", string(evt))
}

func digest(evt json.RawMessage) (out json.RawMessage, err error) {
	var input interface{}
	err = json.Unmarshal(evt, &input)
	if err != nil {
		return out, err
	}
	log.WithField("input", input).Debug("input")
	if rec, ok := input.(map[string]interface{}); ok {
		for key, val := range rec {
			log.Infof(" [========>] %s = %s", key, val)
			switch key {
			case "firstName", "lastName", "phoneNumber", "name", "moreInfo", "streetAddress", "city", "state":
				if val, ok := val.(string); ok {
					data, err := base64.StdEncoding.DecodeString(val)
					if err != nil {
						log.WithError(err).Debug("ignore not base64")
						data = []byte(val)
					}
					log.WithField("data", string(data)).Debug("decoded")
					rec[key] = string(data)
				}
			}
		}
		out, err = json.Marshal(rec)
	}
	return
}
