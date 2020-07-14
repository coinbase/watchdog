package controller

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/controller/notify"
	"github.com/coinbase/watchdog/primitives/datadog"
	"github.com/coinbase/watchdog/primitives/datadog/pollster"
	"github.com/coinbase/watchdog/primitives/datadog/types"
	"github.com/coinbase/watchdog/primitives/git"
	"github.com/coinbase/watchdog/primitives/github"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// New is a constructor which returns a new instance of Controller.
func New(cfg *config.Config, opts ...Option) (*Controller, error) {
	wc := &Controller{
		cfg: cfg,
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(wc); err != nil {
				return nil, err
			}
		}
	}

	notifySenders := []notify.Sender{notify.NewGithubCommentSender(3, time.Second*5, wc.github)}
	if slackToken := cfg.SystemConfig.GetSlackToken(); slackToken != "" {
		notifySenders = append(notifySenders, notify.NewSlackSender(context.Background(), slackToken))
	}
	wc.notificationHandler = notify.NewHandler(notifySenders...)

	return wc, nil
}

// Controller is the business logic component of watchdog app.
type Controller struct {
	sync.Mutex

	cfg                 *config.Config
	datadog             *datadog.Datadog
	git                 git.Client
	github              github.Client
	pollster            pollster.Pollster
	notificationHandler *notify.Handler
}

// ComponentExists checks if a component file on the master branch.
func (c *Controller) ComponentExists(component types.Component, team, project string, id string) bool {
	c.Lock()
	defer c.Unlock()

	filename := c.cfg.ComponentPath(component, team, project, id)
	err := c.git.PullMaster()
	if err != nil {
		logrus.Errorf("Error checking if component exists, pull master returned error: %s", err)
		return false
	}

	_, err = c.git.ReadFile(filename)
	return err == nil
}

// CreatePullRequest takes a map of datadog components and their ids
// checks for the difference between current state and state from master branch
// and creates a pull requests if needed. This is the main controller's function.
func (c *Controller) CreatePullRequest(team, project, configFile string, componentsMap map[types.Component][]string) error {
	if len(componentsMap) == 0 {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	// remove the leading slash in the beginning of the filename
	configFile = strings.TrimLeft(configFile, "/")

	logrus.Debugf("Start preparing pull request. Team [%s], project [%s], componentsMap [%+v]", team, project, componentsMap)
	err := c.git.PullMaster()
	if err != nil {
		return errors.Wrap(err, "unable to pull git master")
	}

	if team == "" {
		return errors.Errorf("empty team with component map %+v", componentsMap)
	}

	// create a new branch
	branch := fmt.Sprintf("refs/heads/%s/%d", team, time.Now().UnixNano())
	err = c.git.CreateBranch(branch)
	if err != nil {
		return errors.Wrapf(err, "unable to create branch %s", branch)
	}
	defer func() {
		err = c.git.RemoveBranch(branch)
		if err != nil {
			logrus.Errorf("Error removing local branch %s: %s", branch, err)
		}
	}()

	// checkout to a newly created branch
	err = c.git.Checkout(branch, false, false)
	if err != nil {
		return errors.Wrapf(err, "unable to checkout to branch %s", branch)
	}

	// add files from component map to a git commit
	for component, ids := range componentsMap {
		err = c.addFiles(team, project, component, ids)
		if err != nil {
			logrus.Errorf("error adding component %s files: %s", component, err)
		}
	}

	// rely on git status to see if added files are different from master branch
	isClean, patch, err := c.git.Clean()
	if err != nil {
		return errors.Wrap(err, "unable to run git clean")
	}

	// if nothing changed, return
	if isClean {
		logrus.Debugf("No changes found for components: %+v. Skipping", componentsMap)
		return nil
	}

	pullRequestTitle, pullRequestBody := c.preparePullRequestDescription(team, patch, configFile, c.cfg.PullRequestBodyExtra(), componentsMap)

	logrus.Infof("A change has been detected. Patch:\n%s", patch)

	// create a new commit
	msg, commitHash, err := c.git.Commit("Add modified component files")
	if err != nil {
		return errors.Wrap(err, "unable to make a new commit")
	}

	logrus.Debugf("A new commit created %s\n%s", commitHash, msg)

	// find duplicate and outdated PRs
	// duplicates are PRs that have exactly the same change in them, outdated are the opposite
	duplicatePRs, outdatedPRs, err := c.findOpenPRs(pullRequestTitle, commitHash, c.cfg.SystemConfig.GitUser())
	if err != nil {
		return errors.Wrapf(err, "unable to find open PRs")
	}

	// if duplicate PRs are found, exit nothing to do here
	if len(duplicatePRs) > 0 {
		logrus.Infof("Found duplicate PRs: %v", func() (prs []int) {
			for _, duplicatePR := range duplicatePRs {
				prs = append(prs, duplicatePR.Number)
			}
			return
		}())
		return nil
	}

	logrus.Info("No opened PRs found")

	// push changes to remote branch
	err = c.git.Push(branch)
	if err != nil {
		return errors.Wrapf(err, "unable to push changes to remote branch %s", branch)
	}

	// create a new pull request
	newPRNumber, err := c.createNewPullRequest(context.Background(), pullRequestTitle, branch, "master", pullRequestBody)
	if err != nil {
		return errors.Wrapf(err, "unable to create a new pull request")
	}

	// notify slack channel about a new pull request
	c.notify(
		configFile,
		fmt.Sprintf("A new pull request https://%s/%s/%s/pull/%d has been created", c.cfg.GetGithubBaseURL(), c.cfg.GetGithubProjectOwner(), c.cfg.GetGithubRepo(), newPRNumber),
		"")

	// close outdated PRs, do not exit on failure
	c.tryCloseOutdatedPRs(newPRNumber, outdatedPRs)
	return nil
}

func (c *Controller) notify(configFile, title, body string) {
	// TODO: refactor method do be generic for different notification backends.
	userConfig, err := c.cfg.UserConfigFromFile(configFile, false)
	if err == nil {
		err = c.notificationHandler.AddComment(
			context.Background(),
			notify.NInfo,
			title,
			body,
			notify.WithSlackMessage(userConfig.Meta.Slack))
		if err != nil {
			logrus.Errorf("Error adding a notification: %s", err)
		}
	} else {
		logrus.Errorf("Error retrieving user config from a file %s: %s", configFile, err)
	}
}

func (c *Controller) tryCloseOutdatedPRs(newPRNumber int, prs []*github.PullRequest) {
	for _, pr := range prs {
		logrus.Debugf("Closing PR %d branch %s", pr.Number, pr.Branch)
		err := c.closePullRequestRemoveBranch(pr.Number, pr.Branch)
		if err != nil {
			logrus.Errorf("Error closing PR %d: %s", pr.Number, err)
		}

		// add comment to closed PR
		err = c.github.CreatePullRequestComment(context.Background(), pr.Number, fmt.Sprintf(":warning: **Closed in favor of #%d**", newPRNumber))
		if err != nil {
			logrus.Errorf("Error commenting on pull request %d: %s", pr.Number, err)
		}
	}
}

func (c *Controller) preparePullRequestDescription(team, patch, configFile, bodyExtra string, componentsMap map[types.Component][]string) (title, body string) {
	title = fmt.Sprintf("[Automated PR] Update datadog component files owned by [%s] - %s", team, configFile)

	body = "Modified component files have been detected and a new PR has been created\n\n"
	body += "The following components are different from master branch:\n" + patch
	body += "\n\n"

	// if only one component with a single ID, add a component name to title and
	// a warning to body.
	if len(componentsMap) == 1 {
		for name, ids := range componentsMap {
			if len(ids) == 1 {
				title += fmt.Sprintf(" %s %s", name, ids[0])
				body += ":warning: **Closing this PR will revert all changes made in datadog!!!**"
			}
		}
	}

	if bodyExtra != "" {
		body += "\n\n"
		body += bodyExtra
	}

	return
}

func (c *Controller) addFiles(team, project string, component types.Component, ids []string) error {
	for _, id := range ids {
		// build a filepath to component json
		filename := c.cfg.ComponentPath(component, team, project, id)

		// allocate buffer for datadog component
		buf := new(bytes.Buffer)
		err := c.datadog.Write(component, id, buf)
		if err != nil {
			logrus.Errorf("unable to write a component %s with id %s to a buffer: %s", component, id, err)
			continue
		}

		// create a new file on filesystem
		err = c.git.NewFile(filename, buf.Bytes())
		if err != nil {
			return errors.Wrapf(err, "unable to create a new file %s on git workspace", filename)
		}

		err = c.git.Add(filename)
		if err != nil {
			return errors.Wrapf(err, "unable to add a file %s to a commit", filename)
		}
	}

	return nil
}

func (c *Controller) createNewPullRequest(ctx context.Context, title, branch, base, description string) (int, error) {
	output, prNumber, err := c.github.CreatePullRequest(ctx, title, branch, base, description)
	if err != nil {
		return 0, err
	}

	logrus.Infof("Pull request created: %s", output)
	return prNumber, nil
}

func (c *Controller) findOpenPRs(title, newCommitHash, owner string) (duplicates, outdated []*github.PullRequest, err error) {
	// find the similar PRs by a title
	logrus.Infof("Searching open PRs on github with title %s", title)
	prs, err := c.github.FindPullRequests(context.Background(), owner, title)
	if err != nil {
		return nil, nil, err
	}

	for _, pr := range prs {
		logrus.Debugf("Detected the following files in PR: %v", pr.AllFiles())
		differentCommits, patch, err := c.git.DiffCommits(pr.SHA, newCommitHash, pr.AllFiles()...)
		if err != nil {
			return nil, nil, err
		}

		if differentCommits {
			// the commits are different, meaning this PR is outdated
			logrus.Infof("found diff between %s and %s\n%s\n", pr.SHA, newCommitHash, patch)
			logrus.Infof("PR %d will be closed because it is outdated", pr.Number)
			outdated = append(outdated, pr)
		} else {
			// there is no difference between commits for given files, meaning this is a duplicate PR
			duplicates = append(duplicates, pr)
		}
	}

	return
}

func (c *Controller) closePullRequestRemoveBranch(number int, branch string) error {
	err := c.github.ClosePullRequests([]int{number}, true)
	if err != nil {
		return errors.Wrapf(err, "unable to close a pull request %d", number)
	}

	return nil
}

func (c *Controller) error(errs []string) error {
	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "; "))
}
