package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

const (
	defaultGithubBaseURL    = "github.com"
	defaultGithubAPIBaseURL = "api.github.com"
	rsaPrivateKeyType       = "RSA PRIVATE KEY"
)

var (
	// ErrInvalidPrivateKey is returned if the given private key is not valid.
	ErrInvalidPrivateKey = errors.New("invalid private key")
)

// SystemConfig is an interface which describes the application system config.
// System config is a basic parameters (required or optional) for running the service.
// In this config we use use git command to create commits, push to remote branch as well as github API
// to open/manage pull requests, reviews etc.
type SystemConfig interface {
	// config accessors
	GetDatadogDataPath() string
	GetDatadogAPIKey() string
	GetDatadogAPPKey() string
	GetDatadogPollingScheduler() string
	GetDatadogPollingInterval() time.Duration
	GetGithubDatadogDataPath() string
	GetGithubBaseURL() string
	GetGithubProjectOwner() string
	GetGithubRepo() string
	GetGithubIntegrationID() int
	GetGithubAppInstallationID() int
	GetGithubWebhookSecret() string
	GetLoggingLevel() string
	GetLoggingJSON() bool
	GetIgnoreKnownHosts() bool

	// GithubAPIURL returns a path to github API endpoint. This is useful for enterprise github, where API url
	// is different from the github.com.
	GithubAPIURL() string

	// GitURL returns a URL to git repository. The git repository is used to store datadog configs.
	GitURL() string

	// GithubAppPrivateKeyBytes returns a private key in bytes
	GithubAppPrivateKeyBytes() []byte

	// GetHTTPSecret returns a http secret used to auth a user performing API requests.
	GetHTTPSecret() string

	// GetHTTPPort returns a port to listen on.
	GetHTTPPort() int

	GitUser() string
	GitEmail() string

	GetSlackToken() string

	// PullRequestBodyExtra returns a string to append to an automatically created pull request.
	PullRequestBodyExtra() string
}

// NewSystemConfig creates a new instance of system config.
func NewSystemConfig() (SystemConfig, error) {
	cfg := &envVarSysConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse environment variables")
	}

	_, err = cfg.privateKey()
	if err != nil {
		return nil, errors.Wrap(err, "invalid private key, must be RSA PEM")
	}

	return cfg, nil
}

// envVarSysConfig is a config structure for watchdog app.
// It includes options and flags to run the service.
type envVarSysConfig struct {
	// DatadogAPIKey is an API key from datadog.
	DatadogAPIKey string `env:"DD_API_KEY,required"`

	// DatadogAPPKey is an APP key from datadog.
	DatadogAPPKey string `env:"DD_APP_KEY,required"`

	// DatadogPollingScheduler is used to define a datadog polling scheduler.
	// The default is simple pollster.
	DatadogPollingScheduler string `env:"DATADOG_POLLING_SCHEDULER" envDefault:"simple"`

	// DatadogPollingInterval sets an interval for datadog to poll the dashboards/monitors.
	// TODO: This parameter should be a part of datadog simple polling scheduler.
	DatadogPollingInterval time.Duration `env:"DATADOG_POLLING_INTERVAL" envDefault:"20s"`

	// IgnoreKnownHosts is an option to ignore or respect the ssh known hosts when cloning repo over ssh.
	// If set to false, the file from `SSH_KNOWN_HOSTS` env variable will be used.
	// Default to ignore
	IgnoreKnownHosts bool `env:"IGNORE_KNOWN_HOSTS" envDefault:"true"`

	// GithubDatadogDataPath is a parameter used to config a base path for datadog assets like
	// dashboards, monitors etc.
	GithubDatadogDataPath string `env:"GITHUB_ASSETS_STORE_PATH" envDefault:"data"`

	// GithubBaseURL stands for a base github URL. Useful for enterprise github.
	GithubBaseURL string `env:"GITHUB_BASE_URL"`

	// GithubAppPrivateKey is a private key generated from github app.
	GithubAppPrivateKey string `env:"GITHUB_APP_PRIVATE_KEY,required"`

	// GithubProjectOwner is an account on github who owns a repository.
	GithubProjectOwner string `env:"GITHUB_PROJECT_OWNER,required"`

	// GithubRepo is a repository on github used to save dashboards/monitors to.
	GithubRepo string `env:"GITHUB_REPO,required"`

	// GithubIntegrationID is an integration id from github app.
	GithubIntegrationID int `env:"GITHUB_APP_INTEGRATION_ID,required"`

	// GithubAppInstallationID is an installation ID of github app.
	GithubAppInstallationID int `env:"GITHUB_APP_INSTALLATION_ID,required"`

	// GithubWebhookSecret a webhook can be configured with the secret.
	GithubWebhookSecret string `env:"GITHUB_WEBHOOK_SECRET"`

	// LoggingLevel sets a logging level for a given application.
	LoggingLevel string `env:"LOGGING_LEVEL"`

	// LoggingJSON enables the logs in json. Useful for elasticsearch.
	LoggingJSON bool `env:"LOGGING_JSON"`

	// TODO: implement a proper authz
	// HTTPSecret holds a secret a client must include in "Authorization" request header
	// in order to call a REST API to reload a config. If blank anyone can reload a config.
	HTTPSecret string `env:"HTTP_SECRET"`

	// HTTPPort sets an HTTP port to listen on
	HTTPPort int `env:"HTTP_PORT" envDefault:"3000"`

	// SlackToken is a slack API token
	SlackToken string `env:"SLACK_TOKEN"`

	PRBodyExtra string `env:"PR_BODY_TEMPLATE"`
}

func (e envVarSysConfig) GetSlackToken() string {
	return e.SlackToken
}

func (e envVarSysConfig) GetDatadogAPIKey() string {
	return e.DatadogAPIKey
}

func (e envVarSysConfig) GetDatadogAPPKey() string {
	return e.DatadogAPPKey
}

func (e envVarSysConfig) GetDatadogPollingScheduler() string {
	return e.DatadogPollingScheduler
}

func (e envVarSysConfig) GetDatadogPollingInterval() time.Duration {
	return e.DatadogPollingInterval
}

func (e envVarSysConfig) GetGithubDatadogDataPath() string {
	return e.GithubDatadogDataPath
}

func (e envVarSysConfig) GetGithubBaseURL() string {
	return e.GithubBaseURL
}

func (e envVarSysConfig) GetGithubProjectOwner() string {
	return e.GithubProjectOwner
}

func (e envVarSysConfig) GetGithubRepo() string {
	return e.GithubRepo
}

func (e envVarSysConfig) GetGithubIntegrationID() int {
	return e.GithubIntegrationID
}

func (e envVarSysConfig) GetGithubAppInstallationID() int {
	return e.GithubAppInstallationID
}

func (e envVarSysConfig) GetGithubWebhookSecret() string {
	return e.GithubWebhookSecret
}

func (e envVarSysConfig) GetLoggingLevel() string {
	return e.LoggingLevel
}

func (e envVarSysConfig) GetLoggingJSON() bool {
	return e.LoggingJSON
}

func (e envVarSysConfig) GetHTTPSecret() string {
	return e.HTTPSecret
}

func (e envVarSysConfig) GetHTTPPort() int {
	return e.HTTPPort
}

func (e envVarSysConfig) GetIgnoreKnownHosts() bool {
	return e.IgnoreKnownHosts
}

// GetDatadogDataPath returns a GithubDatadogDataPath trimming the slack.
func (e envVarSysConfig) GetDatadogDataPath() string {
	return strings.TrimLeft(e.GithubDatadogDataPath, "/")
}

// GitURL returns a URL to git repository
func (e envVarSysConfig) GitURL() string {
	baseURL := e.GithubBaseURL
	if baseURL == "" {
		baseURL = defaultGithubBaseURL
	}

	if e.GithubProjectOwner == "" || e.GithubRepo == "" {
		return ""
	}

	return fmt.Sprintf("git@%s:%s/%s.git", baseURL, e.GithubProjectOwner, e.GithubRepo)
}

// privateKey returns an instance of *rsa.privateKey.
func (e envVarSysConfig) privateKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(e.GithubAppPrivateKey))
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}

	if block.Type != rsaPrivateKeyType {
		return nil, ErrInvalidPrivateKey
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// GithubAppPrivateKeyBytes returns a private key as a slice of bytes.
func (e envVarSysConfig) GithubAppPrivateKeyBytes() []byte {
	return []byte(e.GithubAppPrivateKey)
}

// GithubAPIURL return a URL to github API.
func (e envVarSysConfig) GithubAPIURL() string {
	if e.GithubBaseURL == defaultGithubBaseURL {
		return fmt.Sprintf("https://%s", defaultGithubAPIBaseURL)
	}

	return fmt.Sprintf("https://%s/api/v3", e.GithubBaseURL)
}

// GitUser returns a git user to make commits as.
func (e envVarSysConfig) GitUser() string {
	return "watchdog[bot]"
}

// GitEmail returns a git email.
func (e envVarSysConfig) GitEmail() string {
	return "watchdog[bot]@users.noreply." + e.GithubBaseURL
}

// PullRequestBodyExtra returns an extra string to append to automated PRs.
func (e envVarSysConfig) PullRequestBodyExtra() string {
	return e.PRBodyExtra
}
