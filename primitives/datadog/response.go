package datadog

import (
	"encoding/json"

	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/types"
)

// Component represents a structure of a watchdog component which holds
// one of datadog component (dashboard, monitor etc.)
type Component struct {
	Type types.Component `json:"type"`

	Dashboard   json.RawMessage                 `json:"dashboard,omitempty"`
	Monitor     *client.MonitorWithDependencies `json:"monitor,omitempty"`
	Downtime    json.RawMessage                 `json:"downtime,omitempty"`
	ScreenBoard json.RawMessage                 `json:"screenboard,omitempty"`
}
