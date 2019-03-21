package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unee-t/env"
)

func init() {
	log.SetHandler(jsonhandler.Default)
}

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

	var dat map[string]interface{}
	if err := json.Unmarshal(evt, &dat); err != nil {
		return "", err
	}

	_, actionType := dat["actionType"].(string)

	if actionType {
		err = actionTypeDB(cfg, evt)
		if err != nil {
			return "", err
		}
	} else {
		err = postChangeMessage(cfg, evt)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Response: %s", resp), nil
}
func actionTypeDB(cfg aws.Config, evt json.RawMessage) (err error) {
	// https://github.com/unee-t/lambda2sns/issues/9

	type actionType struct {
		UnitCreationRequestID int    `json:"unitCreationRequestId"`
		UserCreationRequestID int    `json:"userCreationRequestId"`
		Type                  string `json:"actionType"`
	}

	var act actionType

	if err := json.Unmarshal(evt, &act); err != nil {
		log.WithError(err).Fatal("unable to unmarshall payload")
		return err
	}

	ctx := log.WithFields(log.Fields{
		"type":                  act.Type,
		"unitCreationRequestId": act.UnitCreationRequestID,
		"userCreationRequestId": act.UserCreationRequestID,
	})

	switch act.Type {
	case "CREATE_UNIT":
		if act.UnitCreationRequestID == 0 {
			ctx.Error("missing unitCreationRequestId")
			return fmt.Errorf("missing unitCreationRequestId")
		}
	case "CREATE_USER":
		if act.UserCreationRequestID == 0 {
			ctx.Error("missing userCreationRequestId")
			return fmt.Errorf("missing userCreationRequestId")
		}
	default:
		ctx.Errorf("Unknown type: %s", act.Type)
		return fmt.Errorf("Unknown type: %s", act.Type)
	}

	// Establish connection to DB

	e, err := env.New(cfg)
	if err != nil {
		return err
	}

	DSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/unee_t_enterprise?multiStatements=true&sql_mode=TRADITIONAL&timeout=5s",
		e.GetSecret("LAMBDA_INVOKER_USERNAME"),
		e.GetSecret("LAMBDA_INVOKER_PASSWORD"),
		e.Udomain("auroradb"))

	log.Info("Opening database")
	DB, err := sql.Open("mysql", DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}

	casehost := fmt.Sprintf("https://%s", e.Udomain("case"))
	APIAccessToken := e.GetSecret("API_ACCESS_TOKEN")

	url := casehost + "/api/process-api-payload?accessToken=" + APIAccessToken
	log.Infof("Posting to: %s, payload %s", url, evt)

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

	log.Infof("Response code %d, Body: %s", res.StatusCode, string(resBody))

	var isCreatedByMe int
	// https://github.com/unee-t/lambda2sns/issues/9#issuecomment-474238691
	switch res.StatusCode {
	case http.StatusOK:
		isCreatedByMe = 0
	case http.StatusCreated:
		isCreatedByMe = 1
	default:
		return fmt.Errorf("error from MEFE: %s", string(resBody))
	}

	type creationResponse struct {
		ID        string `json:"id"`
		UnitID    string `json:"unitMongoId"`
		UserID    string `json:"userId"`
		Timestamp string `json:"timestamp"`
	}

	var parsedResponse creationResponse

	if err := json.Unmarshal(resBody, &parsedResponse); err != nil {
		log.WithError(err).Fatal("unable to unmarshall payload")
		return err
	}

	switch act.Type {
	case "CREATE_UNIT":
		parsedResponse.ID = parsedResponse.UnitID
	case "CREATE_USER":
		parsedResponse.ID = parsedResponse.UserID
	}

	ctx = ctx.WithFields(log.Fields{
		"id":               parsedResponse.ID,
		"timestamp":        parsedResponse.Timestamp,
		"is_created_by_me": isCreatedByMe,
	})

	ctx.Info("Gonna call ut_creation_success_mefe_unit_id")

	if parsedResponse.ID == "" {
		ctx.Error("missing ID")
		return fmt.Errorf("Missing ID from MEFE response")
	}

	var filledSQL string
	switch act.Type {
	case "CREATE_UNIT":
		templateSQL := `SET @unit_creation_request_id = '%d';
SET @mefe_unit_id = 'unitMongoId (%s)';
SET @creation_datetime = 'timestamp (%s)';
SET @is_created_by_me = %d;
CALL ut_creation_success_mefe_unit_id;
}`
		filledSQL = fmt.Sprintf(templateSQL, act.UnitCreationRequestID, parsedResponse.ID, parsedResponse.Timestamp, isCreatedByMe)
	case "CREATE_USER":
		templateSQL := `SET @user_creation_request_id = '%d';
SET @mefe_user_id = 'userId (%s)';
SET @creation_datetime = 'timestamp (%s)';
SET @is_created_by_me = %d;
CALL  ut_creation_success_mefe_user_id;
}`
		filledSQL = fmt.Sprintf(templateSQL, act.UserCreationRequestID, parsedResponse.ID, parsedResponse.Timestamp, isCreatedByMe)
	default:
		return fmt.Errorf("Unknown type: %s, so no SQL template can be inferred", act.Type)
	}

	ctx.Infof("filledSQL: %s", filledSQL)

	_, err = DB.Exec(filledSQL)
	if err != nil {
		ctx.WithError(err).Error("running sql failed")
	}
	return err
}

// For event notifications https://github.com/unee-t/lambda2sns/tree/master/tests/events
func postChangeMessage(cfg aws.Config, evt json.RawMessage) (err error) {
	e, err := env.New(cfg)
	if err != nil {
		return err
	}
	casehost := fmt.Sprintf("https://%s", e.Udomain("case"))
	APIAccessToken := e.GetSecret("API_ACCESS_TOKEN")

	url := casehost + "/api/db-change-message/process?accessToken=" + APIAccessToken
	log.Infof("Posting to: %s, payload %s", url, evt)

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
		log.Errorf("Response code %d, Body: %s", res.StatusCode, string(resBody))
	}
	return err
}

func main() {
	lambda.Start(handler)
}
