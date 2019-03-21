package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/mnaboka/ghinstallation"
)

// Option is a functional parameter
type Option func(gh *Github) error

// WithJWTTransport configures HTTP transport to use installations token based on JWT.
func WithJWTTransport(owner, repo, githubAPI string, integrationID, installationID int, privateKeyBody []byte) Option {
	return func(gh *Github) error {
		tr := http.DefaultTransport
		itr, err := ghinstallation.New(tr, integrationID, installationID, privateKeyBody, ghinstallation.WithEnterpriseGithub(githubAPI))
		if err != nil {
			return err
		}

		gh.client, err = github.NewEnterpriseClient(githubAPI, "", &http.Client{Transport: itr})
		if err != nil {
			return err
		}
		gh.owner = owner
		gh.repositoryName = repo

		return nil
	}
}
