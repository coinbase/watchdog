package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// DashboardsResponse represents a response by GetDashboards method.
type DashboardsResponse json.RawMessage

// GetModifiedIDsWithin returns a list of IDs that were modified within the given interval.
func (dr DashboardsResponse) GetModifiedIDsWithin(interval time.Duration, fn func(time.Time) time.Duration) ([]int, error) {
	if fn == nil {
		fn = time.Since
	}

	var resp struct {
		Dashes []struct {
			ID       string `json:"id"`
			Modified string `json:"modified"`
		}
	}

	err := json.Unmarshal(dr, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboards response")
	}

	var ids []int
	for _, d := range resp.Dashes {
		if d.Modified == "" {
			return nil, fmt.Errorf("empty modified field, full response: %+v", resp)
		}

		t, err := time.Parse(time.RFC3339Nano, d.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse modified field %s", d.Modified)
		}

		id, err := strconv.Atoi(d.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to convert %s to integer", d.ID)
		}

		if fn(t) < interval {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// GetDashboard returns a raw json of dashboard.
func (c Client) GetDashboard(id int) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%d", dashboardType, id), nil)
	if err != nil {
		return nil, err
	}

	return c.stripJSONFields(resp, c.removeDashboardFields)
}

// UpdateDashboard updates the dashboard from json raw message.
func (c Client) UpdateDashboard(dash json.RawMessage) error {
	dashboard := struct {
		Dash json.RawMessage `json:"dash"`
	}{}

	err := json.Unmarshal(dash, &dashboard)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal dashboard, field dash")
	}

	err = c.genericUpdate(dashboardType, dashboard.Dash)
	if err != nil {
		if err == errInvalidComponent {
			return ErrInvalidDashboard
		}

		return err
	}

	return nil
}

// GetDashboards returns a list of dashboards.
func (c Client) GetDashboards() (DashboardsResponse, error) {
	return c.do("GET", "dash", nil)
}
