package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetAlert(t *testing.T) {
	alertJSON := `
{
  "event_object": "123",
  "notify_audit": false,
  "timeout_h": 0,
  "silenced": false,
  "query": "avg()...",
  "message": "test message",
  "id": 123,
  "name": "test name",
  "no_data_timeframe": null,
  "creator": 9888,
  "notify_no_data": false,
  "renotify_interval": 20,
  "state": "OK",
  "escalation_message": "test message",
  "silenced_timeout_ts": null
}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alert/222" {
			t.Fatalf("expect url \"/alert/222\" Got %s", r.URL.Path)
		}

		if r.URL.RawQuery != "api_key=123&application_key=456" {
			t.Fatalf("expect \"api_key=123&application_key=456\" Got %s", r.URL.RawQuery)
		}

		fmt.Fprint(w, alertJSON)
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL

	alert, err := c.GetAlert(222)
	if err != nil {
		t.Fatal(err)
	}

	cmp := bytes.Compare([]byte(alertJSON), alert)
	if cmp != 0 {
		t.Fatalf("expect response %s. Got %s ", alertJSON, string(alert))
	}
}

func TestClient_UpdateAlert(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alert/55" {
			t.Fatalf("expect url /alert/55 Got %s", r.URL.Path)
		}

		if r.Method != "PUT" {
			t.Fatalf("expect method PUT. Got %s", r.Method)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != `{"id":55}` {
			t.Fatalf("expect \"{\"id\": 55}\". Got %s", string(body))
		}
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL
	err = c.UpdateAlert([]byte(`{"id":55}`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetAlerts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alert" {
			t.Fatalf("expect URL /alert Got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Fatalf("expect method GET. Got %s", r.Method)
		}
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL

	_, err = c.GetAlerts()
	if err != nil {
		t.Fatal(err)
	}
}
