package server

import (
	"fmt"
	"net/http"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/controller"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

const (
	// APIPrefix is a base path to restful APIs.
	APIPrefix = APICommonPrefix + "/v1"
)

// New creates a new instance of a Rounter
func New(cfg *config.Config, opts ...Option) (*Router, error) {

	r := &Router{
		cfg:    cfg,
		router: mux.NewRouter(),
	}

	sub := r.router.PathPrefix(APIPrefix).Subrouter()
	// the webhook is protected by github secret
	sub.HandleFunc("/github/ghwebhook", r.handlerGithubWebHook)

	// protect exposed http endpoints with simple secret, the client is supposed to include
	// "Authorization: <secret>" header to access endpoints.
	sub.Handle("/watchdog/config/reload", simpleAuth(cfg.GetHTTPSecret(), http.HandlerFunc(r.reloadConfig))).Methods("POST")
	sub.HandleFunc("/version", r.handlerVersion)

	for _, opt := range opts {
		if opt != nil {
			err := opt(r)
			if err != nil {
				return nil, errors.Wrap(err, "unable to apply parameter")
			}
		}
	}

	return r, nil
}

// Router is an abstraction over mux.Router
type Router struct {
	version   interface{}
	router    *mux.Router
	c         *controller.Controller
	ghWebHook *github.Webhook
	cfg       *config.Config
}

// Start a new HTTP server
func (r *Router) Start() error {
	httpPort := r.cfg.GetHTTPPort()
	logrus.Infof("Starting a new HTTP server on port %d", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r.router)
}
