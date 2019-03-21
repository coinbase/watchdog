package controller

import (
	"encoding/json"
	"testing"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog"
	"github.com/coinbase/watchdog/primitives/datadog/types"
)

func TestPoll(t *testing.T) {
	ddog, err := datadog.New("123", "345", nil, datadog.WithAccessorGetFn(
		types.ComponentDashboard,
		func(id int) (json.RawMessage, error) {
			return nil, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		UserConfig: &fakeUserConfig{},
	}

	c := &Controller{
		cfg:     cfg,
		datadog: ddog,
		git:     &fakeGitClient{},
		github:  &fakeGithubClient{},
	}

	err = c.Poll(nil)
	if err != nil {
		t.Fatal(err)
	}
}
