package pollster

import (
	"context"
	"time"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/types"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewSimplePollster returns an instance of a simple polling scheduler.
func NewSimplePollster(client *client.Client, interval time.Duration, cfg *config.Config, componentFn func(component types.Component, team, project string, id int) bool) Pollster {
	return &simplePoller{
		interval:         interval,
		cfg:              cfg,
		ca:               newComponentAccessors(client),
		componentAllowed: componentFn,
	}
}

type simplePoller struct {
	ca       *componentAccessors
	interval time.Duration
	cfg      *config.Config

	componentAllowed func(component types.Component, team, project string, id int) bool
}

func newComponentAccessors(c *client.Client) *componentAccessors {
	return &componentAccessors{
		getDashboards:   c.GetDashboards,
		getMonitors:     c.GetMonitors,
		getScreenBoards: c.GetScreenboards,
	}
}

// ComponentAccessors is a structure contains functions to query datadog to get a list of components.
type componentAccessors struct {
	getDashboards   func() (client.DashboardsResponse, error)
	getMonitors     func() (client.MonitorsResponse, error)
	getScreenBoards func() (client.ScreenBoardsResponse, error)
}

// Do in implementation of pollster interface.
func (s *simplePoller) Do(ctx context.Context) chan *Response {
	result := make(chan *Response)
	go s.start(ctx, result)
	return result
}

func (s *simplePoller) start(ctx context.Context, result chan *Response) {
	ticker := time.Tick(s.interval)

	logrus.Infof("Start polling with interval %s", s.interval)
	for {
		select {
		case <-ctx.Done():
			logrus.Warn("Shutting down pollster")
			return
		case <-ticker:
			logrus.Debug("Start polling datadog for changes")
			s.poll(result)
		}
	}
}

func (s *simplePoller) poll(result chan *Response) {
	for component, pollFn := range map[types.Component]func() ([]int, error){
		types.ComponentDashboard:   s.pollDashboards,
		types.ComponentMonitor:     s.pollMonitors,
		types.ComponentScreenboard: s.pollScreenBoards,
	} {
		ids, err := pollFn()
		if err != nil {
			logrus.Error(err)
			continue
		}

		s.sendFilteredResponse(component, ids, result)
	}
}

func (s *simplePoller) pollDashboards() ([]int, error) {
	dashboards, err := s.ca.getDashboards()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dashboards")
	}

	return dashboards.GetModifiedIDsWithin(s.interval, nil)
}

func (s *simplePoller) pollMonitors() ([]int, error) {
	monitors, err := s.ca.getMonitors()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get monitors")
	}

	return monitors.GetModifiedIDsWithin(s.interval, nil)
}

func (s *simplePoller) pollScreenBoards() ([]int, error) {
	screenBoards, err := s.ca.getScreenBoards()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get screenboards")
	}

	return screenBoards.GetModifiedIDsWithin(s.interval, nil)
}

func (s *simplePoller) sendFilteredResponse(component types.Component, ids []int, result chan *Response) {
	for _, id := range ids {
		userConfigFiles := s.cfg.UserConfigFilesByComponentID(component, id)

		// send one event per user file
		for _, userConfigFile := range userConfigFiles {
			logrus.Debugf("Detected a change %s id %d", component, id)
			if s.componentAllowed != nil && !s.componentAllowed(component, userConfigFile.Meta.Team, userConfigFile.Meta.Project, id) {
				logrus.Debugf("Change is not allowed. Skipping")
				continue
			}

			result <- &Response{
				UserConfigFile: userConfigFile,
				Component:      component,
				ID:             id,
			}
		}
	}
}
