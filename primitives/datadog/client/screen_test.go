package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScreenBoardsResponse_GetModifiedIDsWithin(t *testing.T) {
	exampleResponse := `
{
  "screenboards": [
    {
      "read_only": false,
      "resource": "/api/v1/screen/111",
      "description": null,
      "title": "title",
      "created": "2018-11-21T21:22:15.148812+00:00",
      "id": 123,
      "created_by": {
        "disabled": false,
        "handle": "test@test.com",
        "name": "Test Foo",
        "is_admin": false,
        "role": null,
        "access_role": "st",
        "verified": true,
        "email": "test@foo.com",
        "icon": "https://secure.gravatar.com/avatar/1234?s=48&d=retro"
      },
      "modified": "2018-11-26T19:04:27.665768+00:00"
    },
    {
      "read_only": false,
      "resource": "/api/v1/screen/222",
      "description": null,
      "title": "test123",
      "created": "2018-08-21T14:47:40.649718+00:00",
      "id": 456,
      "created_by": {
        "disabled": false,
        "handle": "test2@abc.com",
        "name": "foo bar",
        "is_admin": false,
        "role": null,
        "access_role": "st",
        "verified": true,
        "email": "test@foo.com",
        "icon": "https://secure.gravatar.com/avatar/123?s=48&d=retro"
      },
      "modified": "2018-11-22T16:43:47.908031+00:00"
    }
  ]
}

`
	sr := ScreenBoardsResponse(json.RawMessage(exampleResponse))
	fn := func(t time.Time) time.Duration {
		return time.Millisecond
	}

	ids, err := sr.GetModifiedIDsWithin(time.Second, fn)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 2 {
		t.Fatalf("expect 2 ids got %d", len(ids))
	}

	if ids[0] != "123" || ids[1] != "456" {
		t.Fatalf("expect values 123 and 456. Got %v", ids)
	}
}
func TestClient_GetScreenboards(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/screen" {
			t.Fatalf("expect URL /screen Got %s", r.URL.Path)
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

	_, err = c.GetScreenboards()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetScreenboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/screen/25" {
			t.Fatalf("expect URL /screen Got %s", r.URL.Path)
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

	_, err = c.GetScreenboard("25")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_UpdateScreenboard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/screen/25" {
			t.Fatalf("expect URL /screen Got %s", r.URL.Path)
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

	exampleScreen := []byte(`{"id": 25}`)
	err = c.UpdateScreenboard(exampleScreen)
	if err != nil {
		t.Fatal(err)
	}
}
