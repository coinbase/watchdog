package controller

import (
	"github.com/coinbase/watchdog/primitives/datadog/types"
	"testing"
)

func TestController_preparePullRequestDescription(t *testing.T) {
	c := &Controller{}

	components := map[types.Component][]string{
		types.ComponentDashboard: {"1", "2", "3"},
	}

	title, body := c.preparePullRequestDescription("test-team", "patch-string", "test/file1.yml", "bodyExtra", components)

	expectedTitle := "[Automated PR] Update datadog component files owned by [test-team] - test/file1.yml"
	expectedBody := "Modified component files have been detected and a new PR has been created\n\n"
	expectedBody += "The following components are different from master branch:\npatch-string\n\n"
	expectedBody += "\n\nbodyExtra"

	if title != expectedTitle {
		t.Fatalf("expect title %s .Got %s", expectedTitle, title)
	}

	if body != expectedBody {
		t.Fatalf("expect body %s .Got %s", expectedBody, body)
	}

	components = map[types.Component][]string{
		types.ComponentDashboard: {"1"},
	}
	title, body = c.preparePullRequestDescription("test-team", "patch-string", "test/file1.yml", "", components)
	expectedTitle = "[Automated PR] Update datadog component files owned by [test-team] - test/file1.yml dashboard 1"
	expectedBody = "Modified component files have been detected and a new PR has been created\n\n"
	expectedBody += "The following components are different from master branch:\npatch-string\n\n"
	expectedBody += ":warning: **Closing this PR will revert all changes made in datadog!!!**"

	if title != expectedTitle {
		t.Fatalf("expect title %s .Got %s", expectedTitle, title)
	}

	if body != expectedBody {
		t.Fatalf("expect body %s .Got %s", expectedBody, body)
	}
}
