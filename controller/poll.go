package controller

import (
	"context"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog/pollster"
	"github.com/coinbase/watchdog/primitives/datadog/types"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// PollDatadog will start a datadog polling scheduler.
func (c *Controller) PollDatadog(ctx context.Context) {
	result := c.pollster.Do(ctx)
	logrus.Info("Checking datadog assets against master")

	err := c.Poll(c.cfg.UserConfigFiles())
	if err != nil {
		logrus.Errorf("The following errors raised during polling datadog: %s", err)
	}

	go c.startWatcher(ctx, result)
}

// Poll the datadog components
func (c *Controller) Poll(userConfigFiles []*config.UserConfigFile) error {
	var errs []string
	for _, userFile := range userConfigFiles {
		err := c.CreatePullRequest(userFile.Meta.Team, userFile.Meta.Project, userFile.Meta.FilePath, userFile.Components())
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	return c.error(errs)
}

// ReloadUserConfigsAndPoll will reload the user config nad run Poll()
func (c *Controller) ReloadUserConfigsAndPoll(userConfigFiles []*config.UserConfigFile) error {
	if err := c.cfg.Reload(); err != nil {
		return errors.Wrap(err, "unable to reload user config")
	}

	if len(userConfigFiles) == 0 {
		return c.Poll(c.cfg.UserConfigFiles())
	}

	return c.Poll(userConfigFiles)
}

func (c *Controller) startWatcher(ctx context.Context, result chan *pollster.Response) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Shutting down datadog polling")
			return

		case response := <-result:
			err := c.CreatePullRequest(response.UserConfigFile.Meta.Team,
				response.UserConfigFile.Meta.Project, response.UserConfigFile.Meta.FilePath, map[types.Component][]int{response.Component: {response.ID}})
			if err != nil {
				logrus.Errorf("Error creating a new pull request for detected change: %s", err)
			}
		}
	}
}
