package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// DashboardsResponse represents a response by GetDashboards method.
type DashboardsResponse json.RawMessage

// GetModifiedIDsWithin returns a list of IDs that were modified within the given interval.
func (dr DashboardsResponse) GetModifiedIDsWithin(interval time.Duration, fn func(time.Time) time.Duration) ([]string, error) {
	if fn == nil {
		fn = time.Since
	}

	var resp struct {
		Dashboards []struct {
			ID       string `json:"id"`
			Modified string `json:"modified"`
		}
	}

	err := json.Unmarshal(dr, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboards response")
	}

	var ids []string
	for _, d := range resp.Dashboards {
		if d.Modified == "" {
			return nil, fmt.Errorf("empty modified field, full response: %+v", resp)
		}

		t, err := time.Parse(time.RFC3339Nano, d.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse modified field %s", d.Modified)
		}

		if fn(t) < interval {
			ids = append(ids, d.ID)
		}
	}

	return ids, nil
}

// GetDashboard returns a raw json of dashboard.
func (c Client) GetDashboard(id string) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%s", dashboardType, id), nil)
	if err != nil {
		return nil, err
	}

	return c.stripJSONFields(resp, c.removeDashboardFields)
}

// UpdateDashboard updates the dashboard from json raw message.
func (c Client) UpdateDashboard(dash json.RawMessage) error {
	dashboard := struct {
		Dashboard json.RawMessage `json:"dashboard"`
	}{}

	err := json.Unmarshal(dash, &dashboard)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal dashboard, field dashboard")
	}

	content := struct {
		ID string `json:"id"`
	}{}

	err = json.Unmarshal(dashboard.Dashboard, &content)
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal dashboard")
	}

	if content.ID == "" {
		return ErrInvalidDashboard
	}

	_, err = c.do("PUT", fmt.Sprintf("%s/%s", dashboardType, content.ID), bytes.NewReader(dashboard.Dashboard))
	if err != nil {
		return err
	}

	return nil
}

// GetDashboards returns a list of dashboards.
func (c Client) GetDashboards() (DashboardsResponse, error) {
	return c.do("GET", "dashboard", nil)
}
