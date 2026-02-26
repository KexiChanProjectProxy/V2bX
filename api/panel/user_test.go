package panel

import (
	"bytes"
	"encoding/json"
	"testing"

	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"github.com/vmihailenco/msgpack/v5"
)

// TestUserInfoJSONRoundTrip verifies JSON marshal/unmarshal using jsonv2.
func TestUserInfoJSONRoundTrip(t *testing.T) {
	users := []UserInfo{
		{Id: 1, Uuid: "aaa-bbb-ccc", SpeedLimit: 100, DeviceLimit: 3},
		{Id: 2, Uuid: "ddd-eee-fff", SpeedLimit: 0, DeviceLimit: 0},
	}
	b, err := jsonv2.Marshal(users)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got []UserInfo
	if err := jsonv2.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != len(users) {
		t.Fatalf("expected %d users, got %d", len(users), len(got))
	}
	for i := range users {
		if got[i] != users[i] {
			t.Errorf("[%d] expected %+v, got %+v", i, users[i], got[i])
		}
	}
}

// TestUserInfoMsgpackRoundTrip verifies msgpack encode/decode matches JSON semantics.
func TestUserInfoMsgpackRoundTrip(t *testing.T) {
	original := UserListBody{
		Users: []UserInfo{
			{Id: 10, Uuid: "u10-uuid", SpeedLimit: 200, DeviceLimit: 5},
			{Id: 20, Uuid: "u20-uuid", SpeedLimit: 0, DeviceLimit: 0},
		},
	}
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.Encode(&original); err != nil {
		t.Fatalf("msgpack encode: %v", err)
	}

	var decoded UserListBody
	dec := msgpack.NewDecoder(&buf)
	if err := dec.Decode(&decoded); err != nil {
		t.Fatalf("msgpack decode: %v", err)
	}

	if len(decoded.Users) != len(original.Users) {
		t.Fatalf("expected %d users, got %d", len(original.Users), len(decoded.Users))
	}
	for i := range original.Users {
		if decoded.Users[i] != original.Users[i] {
			t.Errorf("[%d] msgpack mismatch: expected %+v, got %+v", i, original.Users[i], decoded.Users[i])
		}
	}
}

// TestUserInfoJSONCompactStreaming verifies the streaming JSON decoder used in GetUserList.
// It simulates a server response with the compact JSON format: {"users":[...]}
func TestUserInfoJSONCompactStreaming(t *testing.T) {
	serverResponse := `{"users":[{"id":1,"uuid":"stream-uuid-1","speed_limit":50,"device_limit":2},{"id":2,"uuid":"stream-uuid-2","speed_limit":0,"device_limit":0}]}`

	r := bytes.NewReader([]byte(serverResponse))
	dec := jsontext.NewDecoder(r)

	// Find "users" key (same logic as GetUserList)
	for {
		tok, err := dec.ReadToken()
		if err != nil {
			t.Fatalf("read token: %v", err)
		}
		if tok.Kind() == '"' && tok.String() == "users" {
			break
		}
	}

	tok, err := dec.ReadToken()
	if err != nil {
		t.Fatalf("read array token: %v", err)
	}
	if tok.Kind() != '[' {
		t.Fatalf("expected '[', got %c", tok.Kind())
	}

	var users []UserInfo
	for dec.PeekKind() != ']' {
		val, err := dec.ReadValue()
		if err != nil {
			t.Fatalf("read value: %v", err)
		}
		var u UserInfo
		if err := jsonv2.Unmarshal(val, &u); err != nil {
			t.Fatalf("unmarshal user: %v", err)
		}
		users = append(users, u)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	want := []UserInfo{
		{Id: 1, Uuid: "stream-uuid-1", SpeedLimit: 50, DeviceLimit: 2},
		{Id: 2, Uuid: "stream-uuid-2", SpeedLimit: 0, DeviceLimit: 0},
	}
	for i := range want {
		if users[i] != want[i] {
			t.Errorf("[%d] expected %+v, got %+v", i, want[i], users[i])
		}
	}
}

// TestUserInfoMsgpackVsJSONEquivalence verifies msgpack and JSON produce the same UserInfo values.
func TestUserInfoMsgpackVsJSONEquivalence(t *testing.T) {
	u := UserInfo{Id: 7, Uuid: "equiv-test", SpeedLimit: 42, DeviceLimit: 1}

	// JSON path
	jb, _ := json.Marshal(u)
	var fromJSON UserInfo
	json.Unmarshal(jb, &fromJSON)

	// Msgpack path
	mb, _ := msgpack.Marshal(u)
	var fromMsgpack UserInfo
	msgpack.Unmarshal(mb, &fromMsgpack)

	if fromJSON != fromMsgpack {
		t.Errorf("JSON and msgpack produce different results: json=%+v msgpack=%+v", fromJSON, fromMsgpack)
	}
}

// TestUserTrafficFields verifies UserTraffic struct fields are intact.
func TestUserTrafficFields(t *testing.T) {
	ut := UserTraffic{UID: 5, Upload: 1024, Download: 2048}
	if ut.UID != 5 || ut.Upload != 1024 || ut.Download != 2048 {
		t.Errorf("UserTraffic fields incorrect: %+v", ut)
	}
}

// TestUserListBodyFields verifies UserListBody can hold multiple users.
func TestUserListBodyFields(t *testing.T) {
	body := UserListBody{
		Users: []UserInfo{
			{Id: 1, Uuid: "a"},
			{Id: 2, Uuid: "b"},
		},
	}
	if len(body.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(body.Users))
	}
}
