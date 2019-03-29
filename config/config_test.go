package config

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coinbase/watchdog/primitives/datadog/types"
)

var privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEArNS1Zv9JxjWyUUzjrxBaYeFULV7v2wjpWXP1yV+KWkZCErXl
6m2KzgmZxwS+19MVBEg7/ML87shz/2T2E72yS8BTPScF3TGB4gGdQBfp3Maul+0t
JiD+5cYxRmgefLvaJW/EJAEQwXI1wUMzDKUyXHastlwt20TVvIyIoUQWlQwVb+gy
McY9ZsrH09z3VriwmeAEcB0qa+ZfkV4NFiTUcjjqsQrdeIzmuYwLzedpDexFOlrz
C6LKSbkQlC1JE67+TUS7tn9Gj9lkjiqOfZ5YMCGhZndyCiieNkgF1AMTl3tUkOd2
XuP7wmGGAjzZFqw8Yhcxe8eqeTGxUFYqGjFl5QIDAQABAoIBAQCe9aDG06SSBk80
0YhUOrE2d13JwRjQl3iwSqRUi2gfsaERvnVx0UCqUlA6qRWyQbWB08JAr0KdiIaP
7tcZvw6e94xXoW2WTPON4Dg2fAgfhCmPGJi/CfgHc+tcO2VXChwQ9KQtDUHQ+m+Q
inMIfWQ9gPVHYK7Yjo4bNhJwaMRwXkGM3lneFgzOMMPbfMz4noWr+4r7c8qseUKu
bLR10ok13r7UICFigcsIh1iKBN5J6eoRIKWe0GzkRv1LjqPwbJxeza17xZNDxZWR
w89UTK1XJU7VrashcU3xQT7SypvgKrGosXVyJTfi+BBLqg7kFgEHb11Yk7kBxdrR
dbEvG5/hAoGBAOSvQRvpxE2cPak8yc+kmsC/GiwThQVWy4SzcVncE49LgfiSNHIL
EJUCSpYNhOrPHa5W5cQmcVNU0GxUEqm7RXhCUus4NBD6jSv+gWnCeV0tbrCplxka
2xj0U6PJemXvtD/ymgDtSxl8I9aYdzBzKbVjub59hE0cDJ2DoiOeEK/dAoGBAMF5
jZa8aUw+kwF2dXfh+ZbgZSCE7fo2JJ9d7v+NgD7ilV/CQ7ydso5cy8T4aQPsJl0M
MUhC0ceYy3dTsin16pNcGfgaj1G6G46C/XkCW+9LLTJB2pkCew5Nugct7VOaCCEW
ChdgltLsudRy3RIHZIV3kOHLhNxBYM4zpCsXVDGpAoGAFyxPK7Xvh3HKqciYJqtm
Zxu2WjsMIrNd4i+Qz+tGLCIZpIekOt42KvNVfYkXK/ga6NyzYcIHf8s7Z47JaVup
uXr3DhDe7c2F2qxqjr3/MFr3OX2l6wxWoVu40gMLnSLCICzEQE3La2Sx+P/wK/+v
fUsCunPboTizao65MmTFCh0CgYEAiLI5N6cnPpd3hjEMDge7ML6atL825PIcLf1Q
P37afZPZti6rbTh+T9eAoUph6EORV2yl5UhQr5VlLIoV90+ozTTlpEYfvL6hea9T
J4xjKE8VP80HhdQa3aBNL4VjiQ3rcHUB7EJyTdSz90awq2xNuX8g/metF3GZ1Bbo
hwmUkwECgYBzJW4ZQwL+KDFeJ+RR5HQcZ6hz3ViHRakZP53YRqO0GQ5m72X1dv3/
0ehO3TsoTuGAJtfOhSYZyoM5yGk+bddUkEgijBfnMNGnlc6YePnqq5mMKW9S7fsr
b4OIrU/K+19UXJxwPg3UDE25zM7tEZk/LbYE1bU9UOFGzhkxa6PIIA==
-----END RSA PRIVATE KEY-----`

func TestDefaultGitURL(t *testing.T) {
	cfg := &envVarSysConfig{
		GithubProjectOwner: "foo",
		GithubRepo:         "bar",
	}

	expectedURL := "git@github.com:foo/bar.git"
	if gitURL := cfg.GitURL(); gitURL != expectedURL {
		t.Fatalf("expect URL %s. Got %s", expectedURL, gitURL)
	}
}

func TestEnterpriseGitURL(t *testing.T) {
	cfg := &envVarSysConfig{
		GithubProjectOwner: "foo",
		GithubRepo:         "bar",
		GithubBaseURL:      "github.company.com",
	}

	expectedURL := "git@github.company.com:foo/bar.git"
	if gitURL := cfg.GitURL(); gitURL != expectedURL {
		t.Fatalf("expect URL %s. Got %s", expectedURL, gitURL)
	}
}

func TestPrivateKey(t *testing.T) {
	cfg := &envVarSysConfig{
		GithubAppPrivateKey: privateKey,
	}

	_, err := cfg.privateKey()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewConfig(t *testing.T) {
	os.Setenv("DD_API_KEY", "123")
	os.Setenv("DD_APP_KEY", "123")
	os.Setenv("GITHUB_APP_PRIVATE_KEY", privateKey)
	os.Setenv("GITHUB_PROJECT_OWNER", "test")
	os.Setenv("GITHUB_REPO", "aaa")
	os.Setenv("GITHUB_APP_INSTALLATION_ID", "22")
	os.Setenv("GITHUB_APP_INTEGRATION_ID", "1")
	os.Setenv("USER_CONFIG_GIT_URL", "1")
	os.Setenv("USER_CONFIG_GIT_PRIVATE_KEY", privateKey)

	_, err := NewConfig(context.Background(), nil, &fakeUserConfig{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindUserConfigs(t *testing.T) {
	cfg := &userGitConfig{
		readDirFn:  ioutil.ReadDir,
		readFileFn: ioutil.ReadFile,
	}
	files, err := cfg.findConfigs("./fixtures/configs")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("expect 3 config files. Got %d", len(files))
	}

	for _, f := range files {
		if !strings.HasPrefix(f.path, "./fixtures/configs/") {
			t.Fatalf("expect a file in fixtures dir. Got %s", f.path)
		}

		if !strings.HasSuffix(f.path, ".yaml") {
			t.Fatalf("expect a yaml file. Got %s", f.path)
		}
	}
}

func TestTeamByID(t *testing.T) {
	cfg := &userGitConfig{
		readDirFn:    ioutil.ReadDir,
		readFileFn:   ioutil.ReadFile,
		basePath:     "./fixtures/configs",
		pullMasterFn: func() error { return nil },
	}
	err := cfg.Reload()
	if err != nil {
		t.Fatal(err)
	}
}

func TestComponentPath(t *testing.T) {
	sysCfg := &fakeSystemsConfig{}
	cfg := &Config{
		SystemConfig: sysCfg,
	}

	expectedPath := "data/foo/bar/test/dashboard-42.json"
	if path := cfg.ComponentPath(types.ComponentDashboard, "foo/bar", "test", 42); path != expectedPath {
		t.Fatalf("expect path %s. Got %s", expectedPath, path)
	}

	expectedPath = "data/infra/sre/screenboard-52.json"
	if path := cfg.ComponentPath(types.ComponentScreenboard, "infra/sre", "", 52); path != expectedPath {
		t.Fatalf("expect path %s. Got %s", expectedPath, path)
	}

	expectedPath = "data/hello/world/monitor-55.json"
	if path := cfg.ComponentPath(types.ComponentMonitor, "hello/world", "", 55); path != expectedPath {
		t.Fatalf("expect path %s. Got %s", expectedPath, path)
	}
}

// mock user config
type fakeUserConfig struct{}

func (f fakeUserConfig) Reload() error {
	return nil
}

func (f fakeUserConfig) UserConfigFilesByComponentID(c types.Component, id int) []*UserConfigFile {
	return nil
}

func (f fakeUserConfig) GetUserConfigBasePath() string {
	return ""
}

func (f fakeUserConfig) UserConfigFiles() []*UserConfigFile {
	return nil
}

func (f fakeUserConfig) UserConfigFromFile(path string, a bool) (*UserConfigFile, error) {
	return nil, nil
}

// mock systems config
type fakeSystemsConfig struct {
}

func (f fakeSystemsConfig) GetDatadogDataPath() string {
	return "data"
}

func (f fakeSystemsConfig) GetDatadogAPIKey() string {
	return ""
}

func (f fakeSystemsConfig) GetDatadogAPPKey() string {
	return ""
}

func (f fakeSystemsConfig) GetDatadogPollingScheduler() string {
	return ""
}

func (f fakeSystemsConfig) GetDatadogPollingInterval() time.Duration {
	return time.Second
}

func (f fakeSystemsConfig) GetGithubDatadogDataPath() string {
	return ""
}

func (f fakeSystemsConfig) GetGithubBaseURL() string {
	return ""
}

func (f fakeSystemsConfig) GetGithubProjectOwner() string {
	return ""
}

func (f fakeSystemsConfig) GetGithubRepo() string {
	return ""
}

func (f fakeSystemsConfig) GetGithubIntegrationID() int {
	return 0
}

func (f fakeSystemsConfig) GetGithubAppInstallationID() int {
	return 0
}

func (f fakeSystemsConfig) GetGithubWebhookSecret() string {
	return ""
}

func (f fakeSystemsConfig) GetLoggingLevel() string {
	return ""
}

func (f fakeSystemsConfig) GetLoggingJSON() bool {
	return false
}

func (f fakeSystemsConfig) GetIgnoreKnownHosts() bool {
	return false
}

func (f fakeSystemsConfig) GithubAPIURL() string {
	return ""
}

func (f fakeSystemsConfig) GitURL() string {
	return ""
}

func (f fakeSystemsConfig) GithubAppPrivateKeyBytes() []byte {
	return nil
}

func (f fakeSystemsConfig) GetHTTPSecret() string {
	return ""
}

func (f fakeSystemsConfig) GetHTTPPort() int {
	return 0
}

func (f fakeSystemsConfig) GitUser() string {
	return ""
}

func (f fakeSystemsConfig) GitEmail() string {
	return ""
}

func (f fakeSystemsConfig) GetSlackToken() string {
	return ""
}

func (f fakeSystemsConfig) PullRequestBodyExtra() string {
	return ""
}
