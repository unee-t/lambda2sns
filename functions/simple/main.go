package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unee-t/env"
)

type withRequestID struct {
	log *log.Entry
}

var account *string
var DB *sql.DB
var APIAccessToken string
var MEFEcase string

func main() {
	log.SetHandler(jsonhandler.Default)

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to load AWS config")
	}

	stssvc := sts.New(cfg)
	input := &sts.GetCallerIdentityInput{}

	req := stssvc.GetCallerIdentityRequest(input)
	result, err := req.Send()
	if err != nil {
		log.WithError(err).Fatal("failed to call stssvc")
	}

	account = result.Account

	e, err := env.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to setup unee-t env")
	}

	DSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/unee_t_enterprise?multiStatements=true&sql_mode=TRADITIONAL&timeout=5s&collation=utf8mb4_unicode_520_ci",
		e.GetSecret("LAMBDA_INVOKER_USERNAME"),
		e.GetSecret("LAMBDA_INVOKER_PASSWORD"),
		e.Udomain("auroradb"))

	DB, err = sql.Open("mysql", DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}
	defer DB.Close()

	MEFEcase = fmt.Sprintf("https://%s", e.Udomain("case"))
	APIAccessToken = e.GetSecret("API_ACCESS_TOKEN")

	lambda.Start(handler)
}

func handler(ctx context.Context, evt json.RawMessage) error {

	c := withRequestID{}
	ctxObj, ok := lambdacontext.FromContext(ctx)
	if ok {
		c.log = log.WithFields(log.Fields{
			"requestID": ctxObj.AwsRequestID,
		})
	} else {
		log.Warn("no requestID context")
	}

	// if err := DB.Ping(); err != nil {
	// 	c.log.WithError(err).Fatal("failed to ping DB")
	// }
	// c.log.WithField("evt", evt).Info("ping")

	var dat map[string]interface{}
	if err := json.Unmarshal(evt, &dat); err != nil {
		return err
	}

	_, actionType := dat["actionType"].(string)

	if actionType {
		c.log.WithField("payload", evt).Info("actionType")
		err := c.actionTypeDB(evt)
		if err != nil {
			// c.log.WithError(err).Error("actionTypeDB")
			return err
		}
	} else {
		c.log.WithField("payload", evt).Info("postChangeMessage")
		err := c.postChangeMessage(evt)
		if err != nil {
			c.log.WithError(err).Error("postChangeMessage")
			return nil // set to nil since we don't want lambda to retry on this type of failure
		}
	}
	return nil
}

func (c withRequestID) actionTypeDB(evt json.RawMessage) (err error) {
	// https://github.com/unee-t/lambda2sns/issues/9

	type actionType struct {
		UnitCreationRequestID       int    `json:"unitCreationRequestId,omitempty"`
		UserCreationRequestID       int    `json:"userCreationRequestId,omitempty"`
		IDmapUserUnitPermissions    int    `json:"idMapUserUnitPermission,omitempty"`
		MEFIRequestID               string `json:"mefeAPIRequestId,omitempty"`
		UpdateUserRequestID         int    `json:"updateUserRequestId,omitempty"`
		UpdateUnitRequestID         int    `json:"updateUnitRequestId,omitempty"`
		RemoveUserFromUnitRequestID int    `json:"removeUserFromUnitRequestId,omitempty"`
		Type                        string `json:"actionType"`
	}

	var act actionType

	if err := json.Unmarshal(evt, &act); err != nil {
		c.log.WithError(err).Fatal("unable to unmarshall payload")
		return err
	}

	ctx := c.log.WithField("actionType", act)
	if act.MEFIRequestID == "" {
		ctx.Error("missing mefeAPIRequestId")
		return fmt.Errorf("missing mefeAPIRequestId")
	}

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
	case "EDIT_UNIT":
		if act.UpdateUnitRequestID == 0 {
			ctx.Error("missing updateUnitRequestId")
			return fmt.Errorf("missing updateUnitRequestId")
		}
	case "CREATE_USER":
		if act.UserCreationRequestID == 0 {
			ctx.Error("missing userCreationRequestId")
			return fmt.Errorf("missing userCreationRequestId")
		}
	case "ASSIGN_ROLE":
		if act.IDmapUserUnitPermissions == 0 {
			ctx.Error("missing idMapUserUnitPermission")
			return fmt.Errorf("missing idMapUserUnitPermission")
		}
	case "DEASSIGN_ROLE":
		if act.RemoveUserFromUnitRequestID == 0 {
			ctx.Error("missing removeUserFromUnitRequestId")
			return fmt.Errorf("missing removeUserFromUnitRequestId")
		}
	default:
		ctx.Error("unknown type")
		return fmt.Errorf("unknown type: %s", act.Type)
	}

	if APIAccessToken == "" {
		ctx.Error("missing API_ACCESS_TOKEN credential")
		return fmt.Errorf("missing API_ACCESS_TOKEN credential")
	}

	url := MEFEcase + "/api/process-api-payload"
	c.log.Debugf("Posting to: %s, payload %s", url, evt)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(evt)))
	if err != nil {
		c.log.WithError(err).Error("constructing POST")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+APIAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.log.WithError(err).Error("POST request")
		return err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.log.WithError(err).Error("failed to read body")
		return err
	}

	var errorMessage string

	var isCreatedByMe int
	// https://github.com/unee-t/lambda2sns/issues/9#issuecomment-474238691
	switch res.StatusCode {
	case http.StatusOK:
		isCreatedByMe = 0
	case http.StatusCreated:
		isCreatedByMe = 1
	default:
		ctx.WithFields(log.Fields{
			"status":   res.StatusCode,
			"evt":      evt,
			"response": string(resBody),
		}).Error("MEFE process-api-payload")
		// We don't stop here since we want to feedback errors to db
		errorMessage = escape(fmt.Sprintf("Error: %s from MEFE: %s, Response: %s from Request: %s", res.Status, url, string(resBody), string(evt)))
	}

	type creationResponse struct {
		ID         string    `json:"id"`
		UnitID     string    `json:"unitMongoId"`
		UserID     string    `json:"userId"`
		Timestamp  time.Time `json:"timestamp"`
		MefeAPIkey string    `json:"mefeApiKey"`
	}

	var parsedResponse creationResponse

	if errorMessage != "" {
		// parsedResponse.ID = fmt.Sprintf("error-%s-%d", act.Type, time.Now().UnixNano())
		ctx = ctx.WithFields(log.Fields{
			"errorMessage": errorMessage,
		})
	} else {
		if err := json.Unmarshal(resBody, &parsedResponse); err != nil {
			c.log.WithError(err).Fatal("unable to unmarshall payload")
			return err
		}

		switch act.Type {
		case "CREATE_UNIT":
			parsedResponse.ID = parsedResponse.UnitID
		case "CREATE_USER":
			parsedResponse.ID = parsedResponse.UserID
		}
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
SET @mefe_api_error_message = '%s';
CALL ut_creation_unit_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL,
			act.UnitCreationRequestID,
			parsedResponse.ID,
			parsedResponse.Timestamp.Format(sqlTimeLayout),
			isCreatedByMe,
			errorMessage)
	case "CREATE_USER":
		templateSQL := `SET @user_creation_request_id = %d;
SET @mefe_user_id = '%s';
SET @creation_datetime = '%s';
SET @is_created_by_me = %d;
SET @mefe_api_error_message = '%s';
SET @mefe_user_api_key = '%s';
CALL ut_creation_user_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL,
			act.UserCreationRequestID,
			parsedResponse.ID,
			parsedResponse.Timestamp.Format(sqlTimeLayout),
			isCreatedByMe,
			errorMessage,
			parsedResponse.MefeAPIkey,
		)
	case "ASSIGN_ROLE":
		templateSQL := `SET @id_map_user_unit_permissions = %d;
SET @creation_datetime = '%s';
SET @mefe_api_error_message = '%s';
CALL ut_creation_user_role_association_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL, act.IDmapUserUnitPermissions, parsedResponse.Timestamp.Format(sqlTimeLayout), errorMessage)
	case "EDIT_USER":
		templateSQL := `SET @update_user_request_id = %d;
SET @updated_datetime = '%s';
SET @mefe_api_error_message = '%s';
CALL ut_update_user_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL, act.UpdateUserRequestID, parsedResponse.Timestamp.Format(sqlTimeLayout), errorMessage)
	case "EDIT_UNIT":
		templateSQL := `SET @update_unit_request_id = %d;
SET @updated_datetime = '%s';
SET @mefe_api_error_message = '%s';
CALL ut_update_unit_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL, act.UpdateUnitRequestID, parsedResponse.Timestamp.Format(sqlTimeLayout), errorMessage)
	case "DEASSIGN_ROLE":
		templateSQL := `SET @remove_user_from_unit_request_id = %d;
SET @updated_datetime = '%s';
SET @mefe_api_error_message = '%s';
CALL ut_remove_user_role_association_mefe_api_reply;`
		filledSQL = fmt.Sprintf(templateSQL, act.RemoveUserFromUnitRequestID, parsedResponse.Timestamp.Format(sqlTimeLayout), errorMessage)
	default:
		return fmt.Errorf("Unknown type: %s, so no SQL template can be inferred", act.Type)
	}
	_, err = DB.Exec(filledSQL)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			// https://github.com/unee-t/lambda2sns/issues/20
			ctx.WithError(err).WithField("sql", filledSQL).Warn("Duplicate entry")
			return nil
		}
		ctx.WithError(err).WithField("sql", filledSQL).Error("running sql failed")
		// automatically retry the invocation twice, with delays between retries
		// https://docs.aws.amazon.com/lambda/latest/dg/retries-on-errors.html
		return err
	}
	// c.log.WithField("stats", DB.Stats()).Info("exec sql")
	if errorMessage != "" && res.StatusCode >= 500 {
		// Payload is valid, but the action took took long (POST time out, database time out)
		return errors.New(errorMessage)
	}
	if errorMessage != "" {
		// Assuming Payload is wrong
		ctx.WithField("status", res.StatusCode).Warn("not returning an error for triggering a retry")
	}
	return err
}

// For event notifications https://github.com/unee-t/lambda2sns/tree/master/tests/events
func (c withRequestID) postChangeMessage(evt json.RawMessage) (err error) {
	url := MEFEcase + "/api/db-change-message/process"
	c.log.Infof("Posting to: %s, payload %s", url, evt)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(evt)))
	if err != nil {
		c.log.WithError(err).Error("constructing POST")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+APIAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.log.WithError(err).Error("POST request")
		return err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.log.WithError(err).Error("failed to read body")
		return err
	}
	if res.StatusCode == http.StatusOK {
		c.log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(resBody),
		}).Info("OK")
	} else {
		c.log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(resBody),
		}).Error("MEFE db-change-message/process")
		return fmt.Errorf("/api/db-change-message/process response code %d, Request: %s Response: %s", res.StatusCode, evt, string(resBody))
	}
	return err
}

// from https://github.com/golang/go/issues/18478#issuecomment-357285669
func escape(source string) string {
	var j int
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = '\r'
			break
		case '\n':
			flag = true
			escape = '\n'
			break
		case '\\':
			flag = true
			escape = '\\'
			break
		case '\'':
			flag = true
			escape = '\''
			break
		case '"':
			flag = true
			escape = '"'
			break
		case '\032':
			flag = true
			escape = 'Z'
			break
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}
