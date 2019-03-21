package controller

import (
	"testing"

	"github.com/coinbase/watchdog/config"
)

func TestFilter(t *testing.T) {
	cfg := &config.Config{
		UserConfig:   &fakeUserConfig{},
		SystemConfig: &fakeSystemsConfig{},
	}

	c := &Controller{
		cfg:    cfg,
		git:    &fakeGitClient{},
		github: &fakeGithubClient{},
	}

	cfgFiles := []string{"cfg/test/foo", "config/123.yaml", "cfg/test/bar.yaml", "cfg/var/test.yml", "config/one/two/three/dashboards.yml"}
	result := c.filterConfigFiles(cfgFiles)
	if len(result) != 2 {
		t.Fatalf("expect 2 ites. Got %v", result)
	}

	if result[0] != "config/123.yaml" {
		t.Fatalf("must contain file cfg/123.yaml. Got %v", result)
	}

	if result[1] != "config/one/two/three/dashboards.yml" {
		t.Fatalf("must contain file cfg/one/two/three/dashboards.yml. Got %v", result)
	}

	componentFiles := []string{"data/team/1/dashboard-123", "some/foo/bar-123"}
	result = c.filterComponentFiles(componentFiles)
	if len(result) != 1 {
		t.Fatalf("expect 1 item. Got %v", result)
	}

	if result[0] != "data/team/1/dashboard-123" {
		t.Fatalf("must contain file data/team/1/dashboard-123. Got %v", result)
	}
}

func TestPullRequestFiles(t *testing.T) {
	cfg := &config.Config{
		UserConfig:   &fakeUserConfig{},
		SystemConfig: &fakeSystemsConfig{},
	}

	c := &Controller{
		cfg:    cfg,
		git:    &fakeGitClient{},
		github: &fakeGithubClient{},
	}

	componentFiles, configFiles, err := c.pullRequestFiles(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(componentFiles) != 1 {
		t.Fatalf("expect 1 component file. Got %v", componentFiles)
	}

	if componentFiles[0] != "data/team/dashboard-123" {
		t.Fatalf("expect component file data/team/dashboard-123. Got %s", configFiles[0])
	}

	if len(configFiles) != 3 {
		t.Fatalf("expect 3 config files. Got %v", configFiles)
	}

	if configFiles[0] != "config/team/dashboards.yml" {
		t.Fatalf("expect file config/team/dashboards.yml. Got %s", configFiles[0])
	}

	if configFiles[1] != "config/team2/monitors.yaml" {
		t.Fatalf("expect file config/team2/monitors.yaml. Got %s", configFiles[1])
	}

	if configFiles[2] != "config/foo/bar/monitor.yml" {
		t.Fatalf("expect file config/foo/bar/monitor.yml. Got %s", configFiles[2])
	}
}
