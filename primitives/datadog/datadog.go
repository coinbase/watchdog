package datadog

import (
	"encoding/json"
	"io"

	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/types"

	"github.com/pkg/errors"
)

var (
	// ErrInvalidComponentTypeID is returned when invalid type ID is used.
	ErrInvalidComponentTypeID = errors.New("invalid component type ID")
)

// New returns a new implementation of datadog APIs.
func New(apiKey, appKey string, c *client.Client, opts ...Option) (*Datadog, error) {
	var err error
	if c == nil {
		c, err = client.New(apiKey, appKey)
		if err != nil {
			return nil, errors.Wrap(err, "unable to configure datadog client")
		}
	}

	dd := &Datadog{
		Client: c,

		getDashboardFn:   c.GetDashboard,
		getMonitorFullFn: c.GetMonitorWithDependencies,
		getDowntimeFn:    c.GetDowntime,
		getScreenBoardFn: c.GetScreenboard,

		updateDashboardFn:   c.UpdateDashboard,
		updateMonitorFn:     c.UpdateMonitorWithDependencies,
		updateDowntimeFn:    c.UpdateDowntime,
		updateScreenBoardFn: c.UpdateScreenboard,
	}

	for _, opt := range opts {
		if opt != nil {
			err := opt(dd)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create a new datadog instance, parameter returned an error")
			}
		}
	}

	return dd, nil
}

// Datadog is a abstraction over datadog api library.
// The abstraction provides simplified interface to query datadog API.
type Datadog struct {
	Client *client.Client

	getDashboardFn   func(string) (json.RawMessage, error)
	getMonitorFullFn func(string, bool) (*client.MonitorWithDependencies, error)
	getDowntimeFn    func(string) (json.RawMessage, error)
	getScreenBoardFn func(string) (json.RawMessage, error)

	updateDashboardFn   func(json.RawMessage) error
	updateMonitorFn     func(*client.MonitorWithDependencies) error
	updateDowntimeFn    func(json.RawMessage) error
	updateScreenBoardFn func(json.RawMessage) error
}

// Write takes a datadog component type ID (dashboard, monitor etc.), id from a datadog and queries the
// corresponding datadog API. The the JSON response will be written to io.Writer.
func (dd *Datadog) Write(component types.Component, id string, to io.Writer) error {
	switch component {
	case types.ComponentDashboard:
		return dd.writeDashboard(id, to)
	case types.ComponentMonitor:
		return dd.writeMonitor(id, to)
	case types.ComponentDowntime:
		return dd.writeDowntime(id, to)
	case types.ComponentScreenboard:
		return dd.writeScreenBoard(id, to)
	}

	return ErrInvalidComponentTypeID
}

func (dd *Datadog) marshalAndWrite(component *Component, to io.Writer) error {
	enc := json.NewEncoder(to)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(component)
}

func (dd *Datadog) writeDashboard(id string, to io.Writer) error {
	dashboard, err := dd.getDashboardFn(id)
	if err != nil {
		return errors.Wrapf(err, "unable to get dashboard %s", id)
	}

	return dd.marshalAndWrite(&Component{
		Type:      types.ComponentDashboard,
		Dashboard: dashboard,
	}, to)
}

func (dd *Datadog) writeMonitor(id string, to io.Writer) error {
	monitor, err := dd.getMonitorFullFn(id, false)
	if err != nil {
		return errors.Wrapf(err, "unable to get monitor %s", id)
	}

	return dd.marshalAndWrite(&Component{
		Type:    types.ComponentMonitor,
		Monitor: monitor,
	}, to)
}

func (dd *Datadog) writeDowntime(id string, to io.Writer) error {
	downtime, err := dd.getDowntimeFn(id)
	if err != nil {
		return errors.Wrapf(err, "unable to get downtime %s", id)
	}

	return dd.marshalAndWrite(&Component{
		Type:     types.ComponentDowntime,
		Downtime: downtime,
	}, to)
}

func (dd *Datadog) writeScreenBoard(id string, to io.Writer) error {
	sb, err := dd.getScreenBoardFn(id)
	if err != nil {
		return errors.Wrapf(err, "unable to get a screenboard %s", id)
	}

	return dd.marshalAndWrite(&Component{
		Type:        types.ComponentScreenboard,
		ScreenBoard: sb,
	}, to)
}

// Update will restore a datadog component from bytes.
func (dd *Datadog) Update(component *Component) error {
	switch component.Type {
	case types.ComponentDashboard:
		return dd.updateDashboard(component.Dashboard)
	case types.ComponentMonitor:
		return dd.updateMonitor(component.Monitor)
	case types.ComponentDowntime:
		return dd.updateDowntime(component.Downtime)
	case types.ComponentScreenboard:
		return dd.updateScreenBoard(component.ScreenBoard)
	}

	return ErrInvalidComponentTypeID
}

func (dd *Datadog) updateDashboard(dashboard json.RawMessage) error {
	if err := dd.updateDashboardFn(dashboard); err != nil {
		return errors.Wrap(err, "unable to update datadog dashboard")
	}

	return nil
}

func (dd *Datadog) updateMonitor(monitor *client.MonitorWithDependencies) error {
	if err := dd.updateMonitorFn(monitor); err != nil {
		return errors.Wrap(err, "unable to update a monitor")
	}

	return nil
}

func (dd *Datadog) updateDowntime(downtime json.RawMessage) error {
	if err := dd.updateDowntimeFn(downtime); err != nil {
		return errors.Wrap(err, "unable to update a downtime")
	}

	return nil
}

func (dd *Datadog) updateScreenBoard(sb json.RawMessage) error {
	if err := dd.updateScreenBoardFn(sb); err != nil {
		return errors.Wrap(err, "unable to update a screenboard")
	}

	return nil
}
