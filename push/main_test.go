package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

var createUnitMessage = `
{
  "city": "",
  "name": "HK_HKG_12ST - 1",
  "type": "Condominium",
  "state": null,
  "country": "Hong Kong",
  "ownerId": "NXnKGEdEwEvMgWQtG",
  "zipCode": null,
  "moreInfo": " ",
  "creatorId": "NXnKGEdEwEvMgWQtG",
  "actionType": "CREATE_UNIT",
  "streetAddress": "12 Staunton Street",
  "mefeAPIRequestId": "e7bb7494-bfa3-11e9-a563-06358cf32556",
  "unitCreationRequestId": 4771
}`

func Test_digest(t *testing.T) {
	var o1 interface{}
	var o2 interface{}

	type args struct {
		evt json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		wantOut json.RawMessage
		wantErr bool
	}{
		{
			name: "arbitary",
			args: args{
				// I don't know the structure of the JSON
				// All I know is that SOMETIMES field values can be base64 encoded
				evt: []byte(`{ "name": "aHR0cHM6Ly9naXRodWIuY29tL3VuZWUtdC9iei1kYXRhYmFzZS9pc3N1ZXMvNzM=" }`),
			},
			wantOut: []byte(`{ "name": "https://github.com/unee-t/bz-database/issues/73" }`),
			wantErr: false,
		},
		{
			name: "no encoding",
			args: args{
				evt: []byte(createUnitMessage),
			},
			wantOut: []byte(createUnitMessage),
			wantErr: false,
		},
		{
			name: "Jožko",
			args: args{
				evt: []byte(`{"streetAddress": "Sm/FvmtvIE1ya3ZpxI1rw6EgMQ==", "name": "Sm/FvmtvIE1ya3ZpxI1rw6EgMg==", "type": "Room"}`),
			},
			wantOut: []byte(`{"streetAddress": "Jožko Mrkvičká 1", "name": "Jožko Mrkvičká 2", "type": "Room"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := digest(tt.args.evt)
			if (err != nil) != tt.wantErr {
				t.Errorf("digest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			json.Unmarshal(gotOut, &o1)
			json.Unmarshal(tt.wantOut, &o2)
			if !reflect.DeepEqual(o1, o2) {
				t.Errorf("digest() = %v, want %v", o1, o2)
			}
		})
	}
}

func Test_id(t *testing.T) {
	type args struct {
		evt json.RawMessage
	}
	tests := []struct {
		name                string
		args                args
		wantDeduplicationId string
		wantGroupId         string
		wantErr             bool
	}{
		{
			name: "notification",
			args: args{
				evt: []byte(`{
            "bz_source_table": "ut_notification_message_new",
            "case_id": 70175,
            "case_reporter_user_id": 6,
            "case_title": "Stains on the side table surface",
            "created_by_user_id": 638,
            "created_datetime": "2019-08-23 07:54:20.000000",
            "current_list_of_invitees": "2, 489",
            "current_resolution": "INVALID",
            "current_severity": "normal",
            "current_status": "RESOLVED",
            "message_truncated": "We removed a user in the role Management Company. This user was un-invited from the case since he has no more role in this unit.",
            "new_case_assignee_user_id": 663,
            "notification_id": "ut_notification_message_new-80675",
            "notification_type": "case_new_message",
            "old_case_assignee_user_id": 663,
            "unit_id": 818
        }`),
			},
			wantDeduplicationId: "ut_notification_message_new-80675",
			wantGroupId:         "notificationType",
			wantErr:             false,
		},
		{
			name: "createUnitMessage",
			args: args{
				evt: []byte(createUnitMessage),
			},
			wantDeduplicationId: "e7bb7494-bfa3-11e9-a563-06358cf32556",
			wantGroupId:         "actionType",
			wantErr:             false,
		},
		{
			name: "error",
			args: args{
				evt: []byte(`not even JSON`),
			},
			wantDeduplicationId: "",
			wantGroupId:         "",
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDeduplicationId, gotGroupId, err := id(tt.args.evt)
			if (err != nil) != tt.wantErr {
				t.Errorf("id() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDeduplicationId != tt.wantDeduplicationId {
				t.Errorf("id() gotDeduplicationId = %v, want %v", gotDeduplicationId, tt.wantDeduplicationId)
			}
			if gotGroupId != tt.wantGroupId {
				t.Errorf("id() gotGroupId = %v, want %v", gotGroupId, tt.wantGroupId)
			}
		})
	}
}
