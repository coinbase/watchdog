package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// ScreenBoardsResponse represents a response from calling /api/v1/dash endpoint
type ScreenBoardsResponse json.RawMessage

// GetModifiedIDsWithin returns a list of screen board IDs if the modified field was changes within the given interval.
func (sr ScreenBoardsResponse) GetModifiedIDsWithin(interval time.Duration, fn func(time.Time) time.Duration) ([]int, error) {
	if fn == nil {
		fn = time.Since
	}

	var resp struct {
		Screenboards []struct {
			ID       int    `json:"id"`
			Modified string `json:"modified"`
		} `json:"screenboards"`
	}

	err := json.Unmarshal(sr, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal monitors response")
	}

	var ids []int
	for _, d := range resp.Screenboards {
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

// GetScreenboards returns a json raw message response from calling /api/v1/dash
func (c Client) GetScreenboards() (ScreenBoardsResponse, error) {
	return c.do("GET", "screen", nil)
}

// UpdateScreenboard updates the screen board from a given raw json message.
func (c Client) UpdateScreenboard(screen json.RawMessage) error {
	err := c.genericUpdate(screenboardType, screen)
	if err != nil {
		if err == errInvalidComponent {
			return ErrInvalidScreenboard
		}

		return err
	}

	return nil
}

// GetScreenboard returns a raw json representation of a screen board.
func (c Client) GetScreenboard(id int) (json.RawMessage, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/%d", screenboardType, id), nil)
	if err != nil {
		return nil, err
	}

	return c.stripJSONFields(resp, c.removeScreenboardFields)
}
