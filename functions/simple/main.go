package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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

var account *string

func init() {
	log.SetHandler(jsonhandler.Default)
}

func reportError(snssvc *sns.SNS, message string) error {
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(fmt.Sprintf("arn:aws:sns:ap-southeast-1:%s:process-api-payload", aws.StringValue(account))),
	})
	_, err := snsreq.Send()
	if err != nil {
		log.Infof("Reported: %s", message)
	}
	return err
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

	account = result.Account

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
			log.WithError(err).Error("actionTypeDB")
			err = reportError(snssvc, err.Error())
			return "", err
		}
	} else {
		err = postChangeMessage(cfg, evt)
		if err != nil {
			log.WithError(err).Error("postChangeMessage")
			err = reportError(snssvc, err.Error())
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
		MEFIRequestID         int    `json:"mefeAPIRequestId"`
		UpdateUserRequestID   int    `json:"updateUserRequestId"`
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
		"updateUserRequestId":   act.UpdateUserRequestID,
	})

	switch act.Type {
	case "CREATE_UNIT":
		if act.UnitCreationRequestID == 0 {
			ctx.Error("missing unitCreationRequestId")
			return fmt.Errorf("missing unitCreationRequestId")
		}
	case "EDIT_USER":
		if act.UpdateUserRequestID == 0 {
			ctx.Error("missing updateUserRequestId")
			return fmt.Errorf("missing updateUserRequestId")
		}
	case "CREATE_USER":
		if act.UserCreationRequestID == 0 {
			ctx.Error("missing userCreationRequestId")
			return fmt.Errorf("missing userCreationRequestId")
		}
	case "ASSIGN_ROLE":
		if act.MEFIRequestID == 0 {
			ctx.Error("missing mefeAPIRequestId")
			return fmt.Errorf("missing mefeAPIRequestId")
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
		return fmt.Errorf("Error: %s from MEFE: %s, Response: %s from Request: %s", res.Status, url, string(resBody), string(evt))
	}

	type creationResponse struct {
		ID        string    `json:"id"`
		UnitID    string    `json:"unitMongoId"`
		UserID    string    `json:"userId"`
		Timestamp time.Time `json:"timestamp"`
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

	// https://dev.mysql.com/doc/refman/8.0/en/datetime.html
	sqlTimeLayout := "2006-01-02 15:04:05"

	var filledSQL string
	switch act.Type {
	case "CREATE_UNIT":
		templateSQL := `SET @unit_creation_request_id = %d;
SET @mefe_unit_id = '%s';
SET @creation_datetime = '%s';
SET @is_created_by_me = %d;
CALL ut_creation_success_mefe_unit_id;`
		filledSQL = fmt.Sprintf(templateSQL, act.UnitCreationRequestID, parsedResponse.ID, parsedResponse.Timestamp.Format(sqlTimeLayout), isCreatedByMe)
	case "CREATE_USER":
		templateSQL := `SET @user_creation_request_id = %d;
SET @mefe_user_id = '%s';
SET @creation_datetime = '%s';
SET @is_created_by_me = %d;
CALL ut_creation_success_mefe_user_id;`
		filledSQL = fmt.Sprintf(templateSQL, act.UserCreationRequestID, parsedResponse.ID, parsedResponse.Timestamp.Format(sqlTimeLayout), isCreatedByMe)
	case "ASSIGN_ROLE":
		templateSQL := `SET @mefe_api_request_id = %d;
SET @creation_datetime = '%s';
CALL ut_creation_success_add_user_to_role_in_unit_with_visibility;`
		filledSQL = fmt.Sprintf(templateSQL, act.MEFIRequestID, parsedResponse.Timestamp.Format(sqlTimeLayout))
	case "EDIT_UDER":
		templateSQL := `SET @update_user_request_id = %d;
SET @updated_datetime = '%s';
CALL ut_update_success_mefe_user;`
		filledSQL = fmt.Sprintf(templateSQL, act.UpdateUserRequestID, parsedResponse.Timestamp.Format(sqlTimeLayout))
	default:
		return fmt.Errorf("Unknown type: %s, so no SQL template can be inferred", act.Type)
	}
	ctx.Infof("DB.Exec filledSQL: %s", filledSQL)
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
