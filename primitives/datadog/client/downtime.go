package client

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// Downtime represents a downtime structure
type Downtime struct {
	MonitorID int    `json:"monitor_id"`
	OrgID     int    `json:"org_id"`
	Disabled  bool   `json:"disabled"`
	Start     uint64 `json:"start"`
	End       uint64 `json:"end"`
	CreatorID int    `json:"creator_id"`
	ID        int    `json:"id"`
	UpdaterID int    `json:"updater_id"`
	Message   string `json:"message"`
}

// Downtimes is a list of downtimes
type Downtimes []*Downtime

// GetByMonitorID returns a downtime by a monitor ID.
func (d Downtimes) GetByMonitorID(id int) *Downtime {
	for _, downtime := range d {
		if downtime.MonitorID == id {
			return downtime
		}
	}

	return nil
}

// GetDowntimes returns a downtimes
func (c Client) GetDowntimes() (Downtimes, error) {
	body, err := c.do("GET", "downtime", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get downtimes")
	}

	downtimes := []*Downtime{}
	err = json.Unmarshal(body, &downtimes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal downtimes")
	}

	return downtimes, nil
}

// GetDowntime returns a downtime
func (c Client) GetDowntime(id int) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%d", downtimeType, id), nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateDowntime updates a downtime.
func (c Client) UpdateDowntime(downtime json.RawMessage) error {
	err := c.genericUpdate(downtimeType, downtime)
	if err != nil {
		if err == errInvalidComponent {
			return ErrInvalidDowntime
		}

		return err
	}

	return nil
}
