package github

import (
	"context"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/waigani/diffparser"
)

// NewGithub returns a new instance fo github object.
func NewGithub(opts ...Option) (Client, error) {
	gh := &Github{}
	for _, opt := range opts {
		if opt != nil {
			err := opt(gh)
			if err != nil {
				return nil, err
			}
		}
	}

	return gh, nil
}

// Github is an abstraction over github API.
type Github struct {
	client         *github.Client
	owner          string
	repositoryName string
}

// PullRequestFiles takes a github diffURL and returns a list of files affected by
// this pull request.
// Example diff url: https://github.com/owner/project/pull/95.diff
func (gh *Github) PullRequestFiles(ctx context.Context, number int) (created, removed, modified []string, err error) {
	patch, _, err := gh.client.PullRequests.GetRaw(ctx, gh.owner, gh.repositoryName, number, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "unable to get raw patch for PR %d", number)
	}

	logrus.Debugf("Detected a patch: %s", patch)

	diff, err := diffparser.Parse(patch)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "unable to parse patch for PR with number %d", number)
	}

	for _, file := range diff.Files {
		// the file was modified
		if file.OrigName == file.NewName {
			modified = append(modified, file.OrigName)
		} else {
			if file.OrigName == "" && file.NewName != "" {
				// if the origin name is empty but new name is not
				// we added a new file.
				created = append(created, file.NewName)
			} else if file.NewName == "" && file.OrigName != "" {
				// if the newname is empty but orig name is not
				// we removed a files
				removed = append(removed, file.OrigName)
			}
		}
	}
	return
}

// CreatePullRequest creates a new pull request.
func (gh *Github) CreatePullRequest(ctx context.Context, title, head, base, body string) (string, int, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(head),
		Base:                github.String(base),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := gh.client.PullRequests.Create(ctx, gh.owner, gh.repositoryName, newPR)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to create a PR")
	}

	return pr.GetHTMLURL(), pr.GetNumber(), nil
}

// ClosePullRequests closes a list of pull requests.
func (gh *Github) ClosePullRequests(prs []int, removeRemoteBranch bool) error {
	for _, pr := range prs {
		pullRequest, _, err := gh.client.PullRequests.Edit(context.Background(), gh.owner, gh.repositoryName, pr, &github.PullRequest{State: github.String("closed")})
		if err != nil {
			return errors.Wrapf(err, "unable to close a pull request %d", pr)
		}
		if removeRemoteBranch {
			ref := "refs/heads/" + pullRequest.GetHead().GetRef()
			logrus.Infof("Removing remote ref %s for PR %d", ref, pr)

			err = gh.RemoveRemoveRef(context.Background(), ref)
			if err != nil {
				logrus.Errorf("Error removing remote ref %s for PR %d: %s", ref, pr, err)
			}
		}
	}

	return nil
}

// RemoveRemoveRef removes a remote reference.
func (gh *Github) RemoveRemoveRef(ctx context.Context, ref string) error {
	_, err := gh.client.Git.DeleteRef(ctx, gh.owner, gh.repositoryName, ref)
	if err != nil {
		return errors.Wrapf(err, "unable to delete ref %s", ref)
	}
	return nil
}

// PullRequest is an abstraction represents a PR.
type PullRequest struct {
	Number    int
	Branch    string
	SHA       string
	CreatedAt *time.Time

	CreatedFiles  []string
	RemovedFiles  []string
	ModifiedFiles []string
}

// AllFiles returns a one slice for all files in pull requested (created, removed and modified)
func (pr PullRequest) AllFiles() []string {
	allFiles := pr.CreatedFiles
	allFiles = append(allFiles, pr.RemovedFiles...)
	allFiles = append(allFiles, pr.ModifiedFiles...)
	return allFiles
}

// FindPullRequests searches and returns a list of pull requests.
func (gh *Github) FindPullRequests(ctx context.Context, owner, titleMatch string) (prs []*PullRequest, err error) {
	pullRequests, _, err := gh.client.PullRequests.List(context.Background(), gh.owner, gh.repositoryName, &github.PullRequestListOptions{
		Head: owner,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list pull requests for user %s", owner)
	}

	for _, pr := range pullRequests {
		if pr.GetTitle() != titleMatch {
			continue
		}

		created, removed, modified, err := gh.PullRequestFiles(ctx, pr.GetNumber())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get pull request %d files", pr.GetNumber())
		}
		logrus.Infof("FindPullRequests found: created %v, removed %v, updated %v", created, removed, modified)

		prs = append(prs, &PullRequest{
			Number:    pr.GetNumber(),
			Branch:    "refs/heads/" + pr.GetHead().GetRef(),
			CreatedAt: pr.CreatedAt,
			SHA:       pr.GetHead().GetSHA(),

			CreatedFiles:  created,
			RemovedFiles:  removed,
			ModifiedFiles: modified,
		})
	}

	return
}

// RequestReviewers add reviewers to a PR.
func (gh *Github) RequestReviewers(pr int, names []string) error {
	_, _, err := gh.client.PullRequests.RequestReviewers(context.Background(), gh.owner, gh.repositoryName, pr, github.ReviewersRequest{
		Reviewers: names,
	})
	return err
}

// CreatePullRequestComment creates a new comment on a given pull request.
func (gh *Github) CreatePullRequestComment(ctx context.Context, id int, text string) error {
	_, _, err := gh.client.Issues.CreateComment(ctx, gh.owner, gh.repositoryName, id, &github.IssueComment{
		Body: github.String(text),
	})
	if err != nil {
		return errors.Wrapf(err, "unable to add a new comment to PR %d", id)
	}

	return nil
}
