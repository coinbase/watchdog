package controller

import (
	"os"
	"time"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog"
	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/pollster"
	"github.com/coinbase/watchdog/primitives/git"
	"github.com/coinbase/watchdog/primitives/github"

	"github.com/pkg/errors"
)

var (
	// ErrDatadogNotInitialized is returned of the datadog is not initialized, but the polling
	// scheduler has been used.
	ErrDatadogNotInitialized = errors.New("datadog not initialized")
)

// Option defines a functional option for Controller.
type Option func(controller *Controller) error

// WithSimplePollster is an option for simple implementation of datadog polling mechanism.
func WithSimplePollster(interval time.Duration, cfg *config.Config) Option {
	return func(wc *Controller) error {
		if wc.datadog == nil {
			return ErrDatadogNotInitialized
		}

		wc.pollster = pollster.NewSimplePollster(wc.datadog.Client, interval, cfg, wc.ComponentExists)
		return nil
	}
}

// WithDatadog is an option used to configure a datadog client.
func WithDatadog(apiKey, appKey string, clientOpts ...client.Option) Option {
	return func(wc *Controller) error {
		var (
			err error
			c   *client.Client
		)

		if len(clientOpts) > 0 {
			c, err = client.New(apiKey, appKey, clientOpts...)
		}

		dd, err := datadog.New(apiKey, appKey, c)
		if err != nil {
			return err
		}
		wc.datadog = dd
		return nil
	}
}

// WithGithub is an option used to configure a github client.
// more about github apps https://developer.github.com/v3/apps/
func WithGithub(owner, repo, githubAPI string, integrationID, installationID int, privateKeyBody []byte) Option {
	return func(wc *Controller) error {
		var err error
		wc.github, err = github.NewGithub(github.WithJWTTransport(owner, repo, githubAPI, integrationID, installationID, privateKeyBody))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithSSHGit is an option used to configure an SSH transport for git.
func WithSSHGit(url, gitUser, gitEmail string, privateKeyBody []byte, ignoreKnownHosts bool) Option {
	return func(wc *Controller) error {
		g, err := git.New(git.WithRSAKey("git", "", privateKeyBody), git.WithIgnoreKnownHosts(ignoreKnownHosts),
			git.WithGitUserEmail(gitUser, gitEmail))
		if err != nil {
			return err
		}

		if err := g.Clone(url, os.Stdout); err != nil {
			return err
		}

		wc.git = g
		return nil
	}
}
