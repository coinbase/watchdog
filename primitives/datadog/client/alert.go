package client

import (
	"encoding/json"
	"fmt"
)

// UpdateAlert updates an alert from raw message.
func (c Client) UpdateAlert(alert json.RawMessage) error {
	err := c.genericUpdate(alertType, alert)
	if err != nil {
		if err == errInvalidComponent {
			return ErrInvalidAlert
		}

		return err
	}

	return nil
}

// GetAlerts returns a list of alerts.
func (c Client) GetAlerts() (json.RawMessage, error) {
	return c.do("GET", "alert", nil)
}

// GetAlert returns an alert with a given ID.
func (c Client) GetAlert(id int) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%d", alertType, id), nil)
	if err != nil {
		return nil, err
	}

	return c.stripJSONFields(resp, c.removeAlertFields)
}
