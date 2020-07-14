package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// MonitorsResponse represents a response from calling /api/v1/monitor endpoint
type MonitorsResponse json.RawMessage

// GetModifiedIDsWithin returns a list of monitor IDs if the modified field was changes within the given interval.
func (mr MonitorsResponse) GetModifiedIDsWithin(interval time.Duration, fn func(time.Time) time.Duration) ([]string, error) {
	if fn == nil {
		fn = time.Since
	}

	var resp []struct {
		ID       int    `json:"id"`
		Modified string `json:"modified"`
	}

	err := json.Unmarshal(mr, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal monitors response")
	}

	var ids []string
	for _, d := range resp {
		if d.Modified == "" {
			return nil, fmt.Errorf("empty modified field, full response: %+v", resp)
		}

		t, err := time.Parse(time.RFC3339Nano, d.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse modified field %s", d.Modified)
		}

		if fn(t) < interval {
			ids = append(ids, strconv.Itoa(d.ID))
		}
	}

	return ids, nil
}

// MonitorWithDependencies represents a monitor with dependencies like Alert and Downtime.
type MonitorWithDependencies struct {
	Monitor  json.RawMessage `json:"monitor"`
	Downtime json.RawMessage `json:"downtime"`
}

// UpdateMonitorWithDependencies will update the dashboard and dependencies.
func (c Client) UpdateMonitorWithDependencies(m *MonitorWithDependencies) error {
	err := c.UpdateMonitor(m.Monitor)
	if err != nil {
		return errors.Wrap(err, "unable to update the monitor")
	}

	// TODO: we are not updating alerts or downtimes on purpose due to a bug
	//       in datadog, which will throw an error anytime we try to update an alert.
	return nil
}

// GetMonitorWithDependencies returns a monitor with dependencies.
func (c Client) GetMonitorWithDependencies(id string, includeDowntime bool) (*MonitorWithDependencies, error) {
	monitor, err := c.GetMonitor(id)
	if err != nil {
		return nil, err
	}

	// alert, err := c.GetAlert(id)
	// if err != nil {
	// 	return nil, err
	// }

	var downtime json.RawMessage

	// try to find a downtime for a given monitor
	if includeDowntime {
		// get a list of all downtimes
		downtimes, err := c.GetDowntimes()
		if err != nil {
			return nil, err
		}

		// find a downtime by a monitor id
		d := downtimes.GetByMonitorID(id)
		if d != nil {
			// if the monitor was found, get the raw message for the full downtime
			downtime, err = c.GetDowntime(strconv.Itoa(d.ID))
		}
	}

	return &MonitorWithDependencies{
		Monitor:  monitor,
		Downtime: downtime,
	}, nil
}

// GetMonitors returns a list of all monitors.
func (c Client) GetMonitors() (MonitorsResponse, error) {
	return c.do("GET", "monitor", nil)
}

// UpdateMonitor updates a monitor from raw message.
func (c Client) UpdateMonitor(monitor json.RawMessage) error {
	err := c.genericUpdate(monitorType, monitor)
	if err != nil {
		if err == errInvalidComponent {
			return ErrInvalidDashboard
		}

		return err
	}

	return nil
}

// GetMonitor returns a monitor by ID
func (c Client) GetMonitor(id string) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%s", monitorType, id), nil)
	if err != nil {
		return nil, err
	}

	return c.stripJSONFields(resp, c.removeMonitorFields)
}
