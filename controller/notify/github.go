package notify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coinbase/watchdog/primitives/github"

	"github.com/pkg/errors"
)

// NewGithubCommentSender constructs an implementation of github comment notification of Sender interface.
func NewGithubCommentSender(maxRetries int, timeout time.Duration, gh github.Client) Sender {
	if gh == nil {
		panic("github parameter was not set")
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	if timeout == 0 {
		timeout = time.Second * 5
	}

	return &githubCommentNotificationWithRetry{
		github:     gh,
		maxRetries: maxRetries,
		timeout:    timeout,
	}
}

// githubCommentNotificationWithRetry is a github comment notification service which implements NotificationSender.
type githubCommentNotificationWithRetry struct {
	pullRequestID int

	maxRetries int
	timeout    time.Duration
	github     github.Client
}

// Notify adds a new comment to github pull request where `id` is a pull request number.
func (g *githubCommentNotificationWithRetry) Notify(ctx context.Context, level NotificationLevel, title, body string) error {
	var errs []string
	for i := 0; i < g.maxRetries; i++ {
		ctx, cancel := context.WithTimeout(ctx, g.timeout)
		err := g.notify(ctx, level, g.pullRequestID, title, body)
		cancel()
		if err != nil {
			errs = append(errs, err.Error())
			time.Sleep(time.Millisecond * 100 * time.Duration(i))
			continue
		}
		return nil
	}

	errStr := strings.Join(errs, "; ")
	return errors.Errorf("reached max retries %d with the following errors: %s", g.maxRetries, errStr)
}

// ID returns a unique sender id.
func (g *githubCommentNotificationWithRetry) ID() SenderID {
	return notifyGithubComment
}

func (g *githubCommentNotificationWithRetry) notify(ctx context.Context, level NotificationLevel, id int, title, body string) error {
	var comment string

	switch level {
	case NSuccess:
		comment = ":white_check_mark: " + title
	case NInfo:
		comment = ":information_source: " + title
	case NWarning:
		comment = fmt.Sprintf(":warning: **%s**", title)
	case NError:
		comment = fmt.Sprintf(":stop_sign: **%s**", title)
	default:
		comment = title
	}

	if body != "" {
		comment += fmt.Sprintf("\n```%s```", body)
	}

	err := g.github.CreatePullRequestComment(ctx, id, comment)
	if err != nil {
		return errors.Wrapf(err, "unable to leave a github comment: %s", comment)
	}

	return nil
}
