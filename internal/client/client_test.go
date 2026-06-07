package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientValidation(t *testing.T) {
	_, err := New("", "token", 30, false, "test")
	if err == nil || !strings.Contains(err.Error(), "endpoint cannot be empty") {
		t.Fatalf("expected empty endpoint error, got: %v", err)
	}

	_, err = New("ftp://example.com", "token", 30, false, "test")
	if err == nil || !strings.Contains(err.Error(), "endpoint must have http or https scheme") {
		t.Fatalf("expected scheme error, got: %v", err)
	}

	_, err = New("https://example.com", "", 30, false, "test")
	if err == nil || !strings.Contains(err.Error(), "token cannot be empty") {
		t.Fatalf("expected empty token error, got: %v", err)
	}
}

func TestClientHeaders(t *testing.T) {
	var receivedReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReq = r
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":1}`), Err: ""})
	}))
	defer server.Close()

	c, err := New(server.URL, "my-secret-token", 30, false, "terraform-provider-nightingale/1.0.0")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = c.doRequest(context.Background(), "POST", "/api/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedReq == nil {
		t.Fatal("no request received")
	}

	if got := receivedReq.Header.Get("X-User-Token"); got != "my-secret-token" {
		t.Errorf("X-User-Token = %q, want %q", got, "my-secret-token")
	}
	if got := receivedReq.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	if got := receivedReq.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
	if got := receivedReq.Header.Get("User-Agent"); got != "terraform-provider-nightingale/1.0.0" {
		t.Errorf("User-Agent = %q, want %q", got, "terraform-provider-nightingale/1.0.0")
	}
}

func TestClientPathBuilding(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
	}))
	defer server.Close()

	cases := []struct {
		endpoint string
		uri      string
		wantPath string
	}{
		{server.URL, "/api/test", "/api/test"},
		{server.URL + "/", "/api/test", "/api/test"},
		{server.URL + "/prefix", "/api/test", "/prefix/api/test"},
		{server.URL + "/prefix/", "/api/test", "/prefix/api/test"},
	}

	for _, tc := range cases {
		c, _ := New(tc.endpoint, "token", 30, false, "test")
		c.doRequest(context.Background(), "GET", tc.uri, nil)

		if gotPath != tc.wantPath {
			t.Errorf("endpoint=%q uri=%q: got path %q, want %q", tc.endpoint, tc.uri, gotPath, tc.wantPath)
		}
	}
}

func TestClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":42}`), Err: ""})
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	env, err := c.doRequest(context.Background(), "GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct{ ID int }
	if err := json.Unmarshal(env.Dat, &result); err != nil {
		t.Fatalf("failed to unmarshal dat: %v", err)
	}
	if result.ID != 42 {
		t.Errorf("ID = %d, want 42", result.ID)
	}
}

func TestClientNon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	_, err := c.doRequest(context.Background(), "GET", "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should contain status code, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain body preview, got: %v", err)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Envelope{Dat: nil, Err: "permission denied"})
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	_, err := c.doRequest(context.Background(), "GET", "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain API error message, got: %v", err)
	}
}

func TestClientMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	_, err := c.doRequest(context.Background(), "GET", "/api/test", nil)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestClientDoesNotLeakToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c, _ := New(server.URL, "super-secret-token", 30, false, "test")
	_, err := c.doRequest(context.Background(), "GET", "/api/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Errorf("error should not leak token: %v", err)
	}
}

func TestClientAlertRuleCRUD(t *testing.T) {
	var lastMethod string
	var lastPath string
	var lastBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path

		if r.Body != nil {
			json.NewDecoder(r.Body).Decode(&lastBody)
		}

		w.WriteHeader(http.StatusOK)
		switch {
		case r.Method == "POST":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"test-rule":""}`), Err: ""})
		case r.Method == "GET" && r.URL.Path == "/api/n9e/busi-group/1/alert-rules":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`[{"id":99,"name":"test-rule","cate":"prometheus","group_id":1,"rule_config":"{\"queries\":[{\"prom_ql\":\"up == 1\"}]}","disabled":0,"severity":2,"create_at":1234567890,"create_by":"admin","update_at":1234567890,"update_by":"admin"}]`), Err: ""})
		case r.Method == "GET":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":99,"name":"updated-rule","cate":"prometheus","group_id":1,"rule_config":"{\"queries\":[{\"prom_ql\":\"up == 1\"}]}","disabled":0,"severity":2,"create_at":1234567890,"create_by":"admin","update_at":1234567890,"update_by":"admin"}`), Err: ""})
		case r.Method == "PUT":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
		case r.Method == "DELETE":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	ctx := context.Background()

	// Create
	rule := &AlertRule{
		Name:           "test-rule",
		DatasourceType: "prometheus",
		GroupID:        1,
		RuleConfig:     FlexibleRuleConfig(`{"queries":[{"prom_ql":"up == 1"}]}`),
		Severity:       2,
	}
	created, err := c.CreateAlertRule(ctx, 1, rule, nil)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID != 99 {
		t.Errorf("created ID = %d, want 99", created.ID)
	}
	if lastMethod != "GET" || lastPath != "/api/n9e/busi-group/1/alert-rules" {
		t.Errorf("create should end with list call: %s %s", lastMethod, lastPath)
	}

	// Read
	got, err := c.GetAlertRule(ctx, 99)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "updated-rule" {
		t.Errorf("got name = %q, want updated-rule", got.Name)
	}
	if lastMethod != "GET" || lastPath != "/api/n9e/alert-rule/99" {
		t.Errorf("get: %s %s", lastMethod, lastPath)
	}

	// Update
	rule.Name = "updated-rule"
	updated, err := c.UpdateAlertRule(ctx, 1, 99, rule, nil)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "updated-rule" {
		t.Errorf("updated name = %q, want updated-rule", updated.Name)
	}
	if lastMethod != "GET" || lastPath != "/api/n9e/alert-rule/99" {
		t.Errorf("update should end with get call: %s %s", lastMethod, lastPath)
	}

	// Delete
	err = c.DeleteAlertRules(ctx, 1, []int64{99})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if lastMethod != "DELETE" || lastPath != "/api/n9e/busi-group/1/alert-rules" {
		t.Errorf("delete: %s %s", lastMethod, lastPath)
	}
	ids, ok := lastBody["ids"].([]interface{})
	if !ok || len(ids) != 1 || fmt.Sprintf("%v", ids[0]) != "99" {
		t.Errorf("delete body ids = %v, want [99]", lastBody["ids"])
	}
}

func TestClientNotifyRuleCRUD(t *testing.T) {
	var lastMethod string
	var lastPath string
	var lastBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		if r.Body != nil {
			json.NewDecoder(r.Body).Decode(&lastBody)
		}

		w.WriteHeader(http.StatusOK)
		switch r.Method {
		case "POST":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`[{"id":11,"name":"test-notify","enable":1}]`), Err: ""})
		case "GET":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":11,"name":"test-notify","enable":1,"user_group_ids":[1],"notify_configs":[{"channel_id":1}]}`), Err: ""})
		case "PUT":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":11,"name":"updated-notify","enable":1}`), Err: ""})
		case "DELETE":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	ctx := context.Background()

	// Create
	rule := &NotifyRule{
		Name:         "test-notify",
		Enable:       1,
		UserGroupIds: []int64{1},
		NotifyConfigs: []NotifyConfig{
			{ChannelID: 1},
		},
	}
	created, err := c.CreateNotifyRule(ctx, rule)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID != 11 {
		t.Errorf("created ID = %d, want 11", created.ID)
	}
	if lastMethod != "POST" || lastPath != "/api/n9e/notify-rules" {
		t.Errorf("create: %s %s", lastMethod, lastPath)
	}

	// Read
	got, err := c.GetNotifyRule(ctx, 11)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "test-notify" {
		t.Errorf("got name = %q", got.Name)
	}
	if lastMethod != "GET" || lastPath != "/api/n9e/notify-rule/11" {
		t.Errorf("get: %s %s", lastMethod, lastPath)
	}

	// Update
	rule.Name = "updated-notify"
	updated, err := c.UpdateNotifyRule(ctx, 11, rule)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "updated-notify" {
		t.Errorf("updated name = %q", updated.Name)
	}
	if lastMethod != "PUT" || lastPath != "/api/n9e/notify-rule/11" {
		t.Errorf("update: %s %s", lastMethod, lastPath)
	}

	// Delete
	err = c.DeleteNotifyRules(ctx, []int64{11})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if lastMethod != "DELETE" || lastPath != "/api/n9e/notify-rules" {
		t.Errorf("delete: %s %s", lastMethod, lastPath)
	}
}

func TestClientAlertSubscribeCRUD(t *testing.T) {
	var lastMethod string
	var lastPath string
	var lastBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		if r.Body != nil {
			json.NewDecoder(r.Body).Decode(&lastBody)
		}

		w.WriteHeader(http.StatusOK)
		switch r.Method {
		case "POST":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":22,"name":"test-sub","group_id":2}`), Err: ""})
		case "GET":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`{"id":22,"name":"test-sub","group_id":2,"disabled":0,"rule_ids":[1],"user_group_ids":[1]}`), Err: ""})
		case "PUT":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
		case "DELETE":
			json.NewEncoder(w).Encode(Envelope{Dat: json.RawMessage(`null`), Err: ""})
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", 30, false, "test")
	ctx := context.Background()

	// Create
	sub := &AlertSubscribe{
		Name:         "test-sub",
		GroupId:      2,
		Disabled:     0,
		RuleIds:      []int64{1},
		UserGroupIds: []int64{1},
	}
	created, err := c.CreateAlertSubscribe(ctx, 2, sub)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID != 22 {
		t.Errorf("created ID = %d, want 22", created.ID)
	}
	if lastMethod != "POST" || lastPath != "/api/n9e/busi-group/2/alert-subscribes" {
		t.Errorf("create: %s %s", lastMethod, lastPath)
	}

	// Read
	got, err := c.GetAlertSubscribe(ctx, 22)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "test-sub" {
		t.Errorf("got name = %q", got.Name)
	}
	if lastMethod != "GET" || lastPath != "/api/n9e/alert-subscribe/22" {
		t.Errorf("get: %s %s", lastMethod, lastPath)
	}

	// Update
	sub.Name = "updated-sub"
	updated, err := c.UpdateAlertSubscribe(ctx, 2, sub)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "updated-sub" {
		t.Errorf("updated name = %q", updated.Name)
	}
	if lastMethod != "PUT" || lastPath != "/api/n9e/busi-group/2/alert-subscribes" {
		t.Errorf("update: %s %s", lastMethod, lastPath)
	}

	// Delete
	err = c.DeleteAlertSubscribes(ctx, 2, []int64{22})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if lastMethod != "DELETE" || lastPath != "/api/n9e/busi-group/2/alert-subscribes" {
		t.Errorf("delete: %s %s", lastMethod, lastPath)
	}
}
