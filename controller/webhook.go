package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/controller/notify"
	"github.com/coinbase/watchdog/primitives/datadog"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

var (
	// ErrInvalidPullRequestPayload is returned if the payload invalid.
	ErrInvalidPullRequestPayload = errors.New("invalid pull request payload")
)

// HandlePullRequestWebhook takes a pull request payload, extracts the files affected by the change
// looks for config files and component files, reloads the user config if such files were affected and restores
// the datadog components if such files were affected.
func (c *Controller) HandlePullRequestWebhook(payload github.PullRequestPayload) error {
	if payload.PullRequest.Number == 0 {
		return ErrInvalidPullRequestPayload
	}

	// account types could be "bot" or "user"
	// user type stands for an account type which opened a pull request
	userType := strings.ToLower(payload.PullRequest.User.Type)

	// sender type stands for an account type which performed an action on pull request e.g. opened/closed
	senderType := strings.ToLower(payload.Sender.Type)

	// merged indicates if the pull request was merged, if false the pull request was closed
	merged := payload.PullRequest.Merged

	prNumber := int(payload.Number)

	// if the pull request was not closed, we should ignore this call.
	if payload.Action != "closed" {
		logrus.Infof("Ignoring a webhook call for PR %d with action %s", prNumber, payload.Action)
		return nil
	}

	// Ignore webhook calls if the watchdog opens or closes PR
	if senderType == "bot" {
		logrus.Infof("Received a webhook call from a bot account %s. Ignoring", payload.Sender.Login)
		return nil
	}

	// 1. if a user created and merged a pull request, watchdog should apply the change from master branch.
	// 2. if a bot created a pull request, but a user closed it (without merging), watchdog should restore from a master branch.
	if (userType == "user" && merged) || (userType == "bot" && !merged) {
		return c.restoreFromFiles(prNumber)
	}

	return nil
}

func (c *Controller) reloadConfigs(prNumber int, userConfigFiles []string) error {
	if len(userConfigFiles) == 0 {
		return nil
	}

	var (
		filesToReload []*config.UserConfigFile
		errs          []string
	)
	for _, userConfigFilePath := range userConfigFiles {
		userCfg, err := c.cfg.UserConfigFromFile(userConfigFilePath, true)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		userCfg.Meta.FilePath = userConfigFilePath
		filesToReload = append(filesToReload, userCfg)
	}

	if err := c.error(errs); err != nil {
		return errors.Wrapf(err, "unable to reload user config")
	}

	logrus.Infof("Pull Request %d has user config files %v. Reloading a config", prNumber, userConfigFiles)

	var (
		title, body string
		level       notify.NotificationLevel = notify.NInfo
	)

	err := c.ReloadUserConfigsAndPoll(filesToReload)
	if err == nil {
		title = "Successfully reloaded user config!"
	} else {
		// not a critical error, do not return
		title = fmt.Sprintf("Error detected while reloading user config. Config files %v; Message: ```%s```", userConfigFiles, err)
		body = fmt.Sprintf("The following errors have been raised: %s", err)
		level = notify.NError
	}

	nonCriticalErr := c.notificationHandler.AddComment(context.Background(), level, title, body, notify.WithGithubPRComment(prNumber))
	if err != nil {
		logrus.Errorf("Error commenting on pull request %d: %s", prNumber, nonCriticalErr)
	}

	return err
}

func (c *Controller) restoreFromFiles(prNumber int) error {
	componentFiles, userConfigFiles, err := c.pullRequestFiles(prNumber)
	if err != nil {
		return errors.Wrapf(err, "unable to extract files from pull request %d", prNumber)
	}

	// find if config files were affected by the change and reload user config if so.
	// the error is not critical so just print out but do not fail.
	err = c.reloadConfigs(prNumber, userConfigFiles)
	if err != nil {
		logrus.Errorf("Unable to reload user config for PR %d: %s", prNumber, err)
	}

	// allow to restore only a single component file per PR
	if len(componentFiles) != 1 {
		return nil
	}

	// restore components
	err = c.restoreDatadogComponents(prNumber, componentFiles)
	if err != nil {
		logrus.Errorf("Error restoring datadog components: %s. Commenting on PR %d", err, prNumber)
		e := c.notificationHandler.AddComment(context.Background(), "ERROR", "Error restoring component", err.Error(), notify.WithGithubPRComment(prNumber))
		if e != nil {
			logrus.Errorf("Error adding a comment to pull request %d: %s", prNumber, err)
		}
	}
	return err

}

func (c *Controller) pullRequestFiles(pullRequestNumber int) (componentFiles, configFiles []string, err error) {
	created, removed, modified, err := c.github.PullRequestFiles(context.Background(), pullRequestNumber)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to find files from pull request %d", pullRequestNumber)
	}

	allFiles := append(created, removed...)
	allFiles = append(allFiles, modified...)

	logrus.Debugf("The following files have been found in pull request %d, created %v, removed %v, modified %v", pullRequestNumber, created, removed, modified)

	// we should filter the following files:
	// for datadog component only the files that were modified. If a new component file was created or a file was removed
	// we should not restore anything.
	// for user config files we should handle all possible scenarios: a config can be added, removed or changed.
	return c.filterComponentFiles(modified), c.filterConfigFiles(allFiles), nil
}

// filter config files which have datadog prefix path
func (c *Controller) filterComponentFiles(files []string) []string {
	filteredFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(file, c.cfg.SystemConfig.GetDatadogDataPath()) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

// filter user config files which have a user config prefix and yml or yaml extension.
func (c *Controller) filterConfigFiles(files []string) []string {
	filteredFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(file, c.cfg.UserConfig.GetUserConfigBasePath()) && (strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, "yml")) {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles
}

// restore the datadog component, do not fail on error for all files.
func (c *Controller) restoreDatadogComponents(prNumber int, componentFiles []string) error {
	c.Lock()
	defer c.Unlock()

	err := c.git.PullMaster()
	if err != nil {
		return errors.Wrap(err, "unable to pull master branch")
	}

	var errs []string
	for _, file := range componentFiles {
		body, err := c.git.ReadFile(file)
		if err != nil {
			return errors.Wrapf(err, "unable to read dashboard %s", file)
		}

		component := &datadog.Component{}
		if err := json.Unmarshal(body, component); err != nil {
			return errors.Wrapf(err, "unable to unmarshal to datadog component. File %s", file)
		}

		logrus.Infof("Restoring datadog component %s from pull request %d", file, prNumber)
		err = c.datadog.Update(component)
		if err == nil {
			e := c.notificationHandler.AddComment(context.Background(), "SUCCESS",
				fmt.Sprintf("Successfully restored %s file %s", component.Type, file), "", notify.WithGithubPRComment(prNumber))
			if e != nil {
				logrus.Errorf("Error commenting on pull request %d: %s", prNumber, err)
			}
		} else {
			errs = append(errs, err.Error())
		}
	}

	return c.error(errs)
}
