package datadog

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/coinbase/watchdog/primitives/datadog/types"
)

func TestDatadogWrite(t *testing.T) {
	dd := &Datadog{
		getDashboardFn: func(id int) (json.RawMessage, error) {
			return []byte(`{"id":2,"title":"test title", "description":"test description"}`), nil
		},
	}

	buf := new(bytes.Buffer)

	err := dd.Write(types.ComponentDashboard, 2, buf)
	if err != nil {
		t.Fatal(err)
	}

	expectedResponse := struct {
		Type      string
		Dashboard struct {
			ID          int
			Description string
			Title       string
		}
	}{}

	err = json.Unmarshal(buf.Bytes(), &expectedResponse)
	if err != nil {
		t.Fatal(err)
	}

	if expectedResponse.Type != "dashboard" {
		t.Fatalf("expected type dashboard. Got %s", expectedResponse.Type)
	}

	if expectedResponse.Dashboard.ID != 2 {
		t.Fatalf("expected dashboard id 2. Got %d", expectedResponse.Dashboard.ID)
	}

	if expectedResponse.Dashboard.Title != "test title" {
		t.Fatalf("expected dashboard title \"test title\". Got %s", expectedResponse.Dashboard.Title)
	}

	if expectedResponse.Dashboard.Description != "test description" {
		t.Fatalf("expected dashboard description \"test description\". Got %s", expectedResponse.Dashboard.Description)
	}
}

func TestDatadogUpdate(t *testing.T) {
	dash := []byte(`{"id":2,"title":"test title","description":"test description"}`)
	dd := &Datadog{
		updateDashboardFn: func(dashboard json.RawMessage) error {
			if cmp := bytes.Compare(dashboard, dash); cmp != 0 {
				t.Fatalf("expecte %s. Got %s", string(dash), string(dashboard))
			}
			return nil
		},
	}

	err := dd.Update(&Component{
		Type:      types.ComponentDashboard,
		Dashboard: dash,
	})

	if err != nil {
		t.Fatal(err)
	}
}
