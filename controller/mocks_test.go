package controller

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog/types"
	"github.com/coinbase/watchdog/primitives/github"
)

// mock user config impl.
type fakeUserConfig struct {
}

func (c fakeUserConfig) UserConfigFilesByComponentID(component types.Component, id int) []*config.UserConfigFile {
	return nil
}

func (c fakeUserConfig) Reload() error {
	return nil
}

func (c fakeUserConfig) GetUserConfigBasePath() string {
	return "config"
}

func (c fakeUserConfig) UserConfigFiles() []*config.UserConfigFile {
	return nil
}

func (c fakeUserConfig) UserConfigFromFile(path string, a bool) (*config.UserConfigFile, error) {
	return nil, nil
}

// mock git impl
type fakeGitClient struct {
}

func (g fakeGitClient) NewFile(name string, body []byte) error {
	return nil
}

func (g fakeGitClient) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (g fakeGitClient) Clone(url string, progress io.Writer) (err error) {
	return nil
}

func (g fakeGitClient) CreateBranch(name string) error {
	return nil
}

func (g fakeGitClient) DiffCommits(commitAHash, commitBHash string, files ...string) (bool, string, error) {
	return false, "", nil
}

func (g fakeGitClient) Checkout(branch string, create, force bool) error {
	return nil
}

func (g fakeGitClient) RemoveBranch(name string) error {
	return nil
}

func (g fakeGitClient) RemoveRemoteBranch(name string) error {
	return nil
}

func (g fakeGitClient) PullMaster() error {
	return nil
}

func (g fakeGitClient) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, nil
}

func (g fakeGitClient) OpenFile(path string) (io.ReadCloser, error) {
	return nil, nil
}

func (g fakeGitClient) Add(path string) error {
	return nil
}

func (g fakeGitClient) Commit(msg string) (string, string, error) {
	return "", "", nil
}

func (g fakeGitClient) Push(branches ...string) error {
	return nil
}

func (g fakeGitClient) Clean() (bool, string, error) {
	return false, "", nil
}

// github mock
type fakeGithubClient struct {
}

func (g fakeGithubClient) PullRequestFiles(ctx context.Context, number int) ([]string, []string, []string, error) {
	created := []string{"config/team/dashboards.yml", "config/team2/monitors.yaml"}
	modified := []string{"data/team/dashboard-123", "config/foo/bar/monitor.yml"}
	return created, nil, modified, nil
}

func (g fakeGithubClient) CreatePullRequest(ctx context.Context, title, head, base, body string) (string, int, error) {
	return "", 0, nil
}

func (g fakeGithubClient) ClosePullRequests(prs []int, removeBranch bool) error {
	return nil
}

func (g fakeGithubClient) FindPullRequests(ctx context.Context, owner, titleMatch string) (prs []*github.PullRequest, err error) {
	return nil, nil
}

func (g fakeGithubClient) RequestReviewers(pr int, names []string) error {
	return nil
}

func (g fakeGithubClient) RemoveRemoveRef(ctx context.Context, ref string) error {
	return nil
}

func (g fakeGithubClient) CreatePullRequestComment(ctx context.Context, id int, text string) error {
	return nil
}

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
