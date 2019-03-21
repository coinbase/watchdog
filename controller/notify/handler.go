package notify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

// Backend is a functional option to configure a sending backend.
type Backend func(handler *Handler) error

// NewHandler returns a new instance of a Handler and initializes backends.
func NewHandler(backends ...Sender) *Handler {
	h := &Handler{
		backend: make(map[SenderID]Sender),
	}

	for _, b := range backends {
		h.backend[b.ID()] = b
	}

	return h
}

// Handler handles notification backends which implement Sender interface.
type Handler struct {
	ctx   context.Context
	level NotificationLevel
	title string
	body  string

	backend map[SenderID]Sender
}

// AddComment adds a new comment to backends.
func (h *Handler) AddComment(ctx context.Context, level NotificationLevel, title, body string, backends ...Backend) error {
	h.ctx = ctx
	h.level = level
	h.title = title
	h.body = body

	var errs []string
	for _, backend := range backends {
		if backend != nil {
			if err := backend(h); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Errorf("The following errors have occurred during notification: %s", strings.Join(errs, "; "))
}

// WithGithubPRComment is a functional option used in AddComment method.
func WithGithubPRComment(pr int) Backend {
	return func(h *Handler) error {
		if pr == 0 {
			return nil
		}

		// get the github comment backend
		// if not found exit with no error
		backend, ok := h.backend[notifyGithubComment]
		if !ok {
			return nil
		}

		githubBackend, ok := backend.(*githubCommentNotificationWithRetry)
		if !ok {
			return errors.New("unable to type assert githubCommentNotificationWithRetry")
		}

		githubBackend.pullRequestID = pr

		defer func() {
			githubBackend.pullRequestID = 0
		}()

		return githubBackend.Notify(h.ctx, h.level, h.title, h.body)
	}
}

// WithSlackMessage is a functional option used in AddComment method to se d notification to slack.
func WithSlackMessage(channel string) Backend {
	return func(h *Handler) error {
		if channel == "" {
			return nil
		}

		backend, ok := h.backend[notifySlackChannel]
		if !ok {
			return nil
		}

		slackBackend, ok := backend.(*slackNotification)
		if !ok {
			return errors.New("unable to type assert slackNotification")
		}

		slackBackend.channel = channel

		// reset the channel
		defer func() {
			slackBackend.channel = ""
		}()

		return slackBackend.Notify(h.ctx, h.level, h.title, h.body)
	}
}
