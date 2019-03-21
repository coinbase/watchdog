package server

import (
	"github.com/coinbase/watchdog/controller"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// Option stands for functional parameter
type Option func(*Router) error

// WithController is an functional parameter to use a controller.
func WithController(wc *controller.Controller) Option {
	return func(r *Router) error {
		r.c = wc
		return nil
	}
}

// WithGithubWebhook is a functional parameter to use a github webhook secret.
func WithGithubWebhook(secret string) Option {
	return func(r *Router) error {
		var err error
		r.ghWebHook, err = github.New(github.Options.Secret(secret))
		if err != nil {
			return errors.Wrap(err, "unable to initialize github webhook")
		}

		return nil
	}
}

// WithVersion sets the version object. This will be used for health endpoint.
func WithVersion(version interface{}) Option {
	return func(r *Router) error {
		if version == nil {
			return errors.New("version cannot be nil")
		}

		r.version = version
		return nil
	}
}
