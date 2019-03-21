package github

import "context"

// Client is a github interface to interact with pull requests.
type Client interface {
	// PullRequestFiles returns a list of files used in a particular pull request.
	PullRequestFiles(ctx context.Context, number int) (created, removed, modified []string, err error)

	// CreatePullRequest creates a new pull request.
	CreatePullRequest(ctx context.Context, title, head, base, body string) (string, int, error)

	// ClosePullRequests closes pull requests.
	ClosePullRequests(prs []int, removeRemoteBranch bool) error

	// FindPullRequests searches pull requests with an owner and title.
	FindPullRequests(ctx context.Context, owner, titleMatch string) (prs []*PullRequest, err error)

	// RequestReviewers assigns the reviewers to a pull request.
	RequestReviewers(pr int, names []string) error

	// RemoveRemoveRef removes a reference from remote git repository.
	RemoveRemoveRef(ctx context.Context, ref string) error

	// CreatePullRequestComment creates a new comment on a pull request.
	CreatePullRequestComment(ctx context.Context, id int, text string) error
}
