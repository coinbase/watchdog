package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var exampleDashboardsRespose = `
{
  "dashboards": [
    {
      "read_only": false,
      "resource": "/api/v1/dash/22",
      "description": "created by foo/bar",
      "title": "test desc",
      "created": "2018-10-22T14:49:07.869620+00:00",
      "id": "my-id-1",
      "created_by": {
        "disabled": false,
        "handle": "foo.bar@test.com",
        "name": "Foo Bar",
        "is_admin": true,
        "role": null,
        "access_role": "adm",
        "verified": true,
        "email": "foo.bar@test.com",
        "icon": "https://secure.gravatar.com/avatar/123?s=48&d=retro"
      },
      "modified": "2018-11-26T16:38:26.588374+00:00"
    },
    {
      "read_only": false,
      "resource": "/api/v1/dash/33",
      "description": "created by foo@coinbase.com",
      "title": "est dashboard",
      "created": "2018-10-20T01:22:03.556752+00:00",
      "id": "my-id-2",
      "created_by": {
        "disabled": false,
        "handle": "foo.bar@test.com",
        "name": "foo/bar",
        "is_admin": true,
        "role": null,
        "access_role": "adm",
        "verified": true,
        "email": "foo.bar@test.com",
        "icon": "https://secure.gravatar.com/avatar/123?s=48&d=retro"
      },
      "modified": "2018-11-21T20:22:57.201615+00:00"
    }
  ]
}
`

func TestClient_UpdateDashboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dashboard/my-test-id" {
			t.Fatalf("expect /dashboard/my-test-id Got %s", r.URL.Path)
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

	exampleDashboard := []byte(`
{
  "dashboard": {
    "id": "my-test-id"
  }
}
`)

	err = c.UpdateDashboard(exampleDashboard)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetDashboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dashboard/other-test-id" {
			t.Fatalf("expect url /dashboard/other-test-id Got %s", r.URL.Path)
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

	_, err = c.GetDashboard("other-test-id")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDashboardsResponse_GetModifiedIDsWithin(t *testing.T) {
	dr := DashboardsResponse(json.RawMessage(exampleDashboardsRespose))
	fn := func(t time.Time) time.Duration {
		return time.Millisecond
	}

	ids, err := dr.GetModifiedIDsWithin(time.Second, fn)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 2 {
		t.Fatalf("expect 2 ids. Got %d", len(ids))
	}

	if ids[0] != "my-id-1" || ids[1] != "my-id-2" {
		t.Fatalf("expect ids my-id-1 and my-id-2. Got %v", ids)
	}
}

func TestClient_GetDashboards(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, exampleDashboardsRespose)
	}))
	defer ts.Close()

	c, err := New("123", "456")
	if err != nil {
		t.Fatal(err)
	}
	c.baseEndpoint = ts.URL

	dashes, err := c.GetDashboards()
	if err != nil {
		t.Fatal(err)
	}

	ids, err := dashes.GetModifiedIDsWithin(time.Second, func(t time.Time) time.Duration {
		return time.Millisecond
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 2 {
		t.Fatalf("expect 2 ids. Got %d", len(ids))
	}

	if ids[0] != "my-id-1" || ids[1] != "my-id-2" {
		t.Fatalf("expect ids my-id-1 and my-id-2. Got %v", ids)
	}
}
