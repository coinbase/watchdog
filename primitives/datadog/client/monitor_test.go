package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitorsResponse_GetModifiedIDsWithin(t *testing.T) {
	exampleMonotorsResponse := `
[
  {
    "tags": [],
    "deleted": null,
    "query": "\"ntp.in_sync\".over(\"*\").last(2).count_by_status()",
    "message": "test",
    "matching_downtimes": [],
    "id": 55,
    "multi": true,
    "name": "[Auto] Clock in sync with NTP",
    "created": "2017-12-18T20:52:20.880245+00:00",
    "created_at": 1513630340000,
    "creator": {
      "id": 716667,
      "handle": "test",
      "name": "Datadog Support",
      "email": "support-user-prod@datadoghq.com"
    },
    "org_id": 156727,
    "modified": "2017-12-18T20:52:20.880245+00:00",
    "overall_state_modified": "2018-11-19T00:16:16.066945+00:00",
    "overall_state": "OK",
    "type": "service check",
    "options": {
      "thresholds": {
        "warning": 1,
        "ok": 1,
        "critical": 1
      },
      "silenced": {}
    }
  },
  {
    "tags": [],
    "deleted": null,
    "query": "avg(last_1h):anomalies...",
    "message": "@slack-",
    "matching_downtimes": [],
    "id": 66,
    "multi": true,
    "name": "monitor name",
    "created": "2018-03-19T19:46:30.093062+00:00",
    "created_at": 1521488790000,
    "creator": {
      "id": 1,
      "handle": "foo.bar@test.com",
      "name": "Foo bar",
      "email": "test.test@test.com"
    },
    "org_id": 156727,
    "modified": "2018-09-10T16:45:06.427714+00:00",
    "overall_state_modified": "2018-11-21T02:58:45.634625+00:00",
    "overall_state": "OK",
    "type": "query alert",
    "options": {
      "notify_audit": false,
      "locked": true,
      "timeout_h": 0,
      "silenced": {},
      "include_tags": true,
      "no_data_timeframe": 10,
      "new_host_delay": 300,
      "require_full_window": true,
      "notify_no_data": true,
      "renotify_interval": 0,
      "escalation_message": "",
      "threshold_windows": {
        "recovery_window": "last_5m",
        "trigger_window": "last_5m"
      },
      "thresholds": {
        "critical": 0.75,
        "critical_recovery": 0
      }
    }
  }
]
`

	mr := MonitorsResponse(json.RawMessage(exampleMonotorsResponse))
	fn := func(t time.Time) time.Duration {
		return time.Millisecond
	}

	ids, err := mr.GetModifiedIDsWithin(time.Second, fn)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 2 {
		t.Fatalf("expect 2 ids. Got %v", len(ids))
	}

	if ids[0] != "55" || ids[1] != "66" {
		t.Fatalf("expect ids 55 and 66 got %v", ids)
	}
}

func TestClient_GetMonitors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/monitor" {
			t.Fatalf("expect URL /monitor Got %s", r.URL.Path)
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

	_, err = c.GetMonitors()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetMonitor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/monitor/25" {
			t.Fatalf("expect URL /monitor Got %s", r.URL.Path)
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

	_, err = c.GetMonitor("25")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateMonitor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/monitor/25" {
			t.Fatalf("expect URL /monitor Got %s", r.URL.Path)
		}

		if r.Method != "PUT" {
			t.Fatalf("expect method PUT. Got %s", r.Method)
		}
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL

	exampleMonitor := []byte(`{"id": 25}`)
	err = c.UpdateMonitor(exampleMonitor)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_MonitorWithDependencies(t *testing.T) {
	monitor := `{"id":1}`
	downtimes := []*Downtime{
		&Downtime{
			ID:        2,
			MonitorID: 1,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/monitor/1":
				fmt.Fprint(w, monitor)
			case "/downtime":
				err := json.NewEncoder(w).Encode(downtimes)
				if err != nil {
					t.Fatal(err)
				}
			case "/downtime/2":
				fmt.Fprint(w, `{"id":3}`)
			default:
				t.Fatalf("invalid URL %s", r.URL.Path)
			}
		} else if r.Method == "PUT" {
			switch r.URL.Path {
			case "/downtime/3":
			case "/monitor/1":
			default:
				t.Fatalf("invalid URL %s", r.URL.Path)
			}
		} else {
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL

	m, err := c.GetMonitorWithDependencies("1", true)
	if err != nil {
		t.Fatal(err)
	}

	if monitor != string(m.Monitor) {
		t.Fatalf("expect %s. Got %s", monitor, string(m.Monitor))
	}

	if `{"id":3}` != string(m.Downtime) {
		t.Fatalf("expect {\"id\":3}. Got %s", string(m.Downtime))
	}

	err = c.UpdateMonitorWithDependencies(m)
	if err != nil {
		t.Fatal(err)
	}
}
