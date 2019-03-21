package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

func (r *Router) handlerGithubWebHook(w http.ResponseWriter, req *http.Request) {
	payload, err := r.ghWebHook.Parse(req, github.PullRequestEvent, github.PingEvent)
	if err != nil {
		msg := "error parsing pull request event " + err.Error()
		logrus.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	switch payload.(type) {
	case github.PingPayload:
		fmt.Fprintf(w, "OK")
		return
	case github.PullRequestPayload:
		pr := payload.(github.PullRequestPayload)
		err = r.c.HandlePullRequestWebhook(pr)
		if err != nil {
			logrus.Errorf("Error handling pull request payload: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (r *Router) reloadConfig(w http.ResponseWriter, req *http.Request) {
	fn := r.c.ReloadUserConfigsAndPoll

	if sync := req.URL.Query().Get("sync"); sync == "1" {
		if err := fn(nil); err != nil {
			logrus.Errorf("Error polling user config: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	go func() {
		if err := fn(nil); err != nil {
			logrus.Errorf("Error reloading user config from web handler: %s", err)
		}
	}()
}

func (r *Router) handlerVersion(w http.ResponseWriter, req *http.Request) {
	if r.version == nil {
		logrus.Warn("Version was not set")
		http.Error(w, "version unset", http.StatusInternalServerError)
		return
	}

	ver := map[string]interface{}{
		"version": r.version,
	}

	if err := json.NewEncoder(w).Encode(ver); err != nil {
		logrus.Errorf("Error encoding a version object")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
