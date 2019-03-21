package pollster

import (
	"context"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog/types"
)

// Response is a response structure returned by a Do method of Pollster interface.
type Response struct {
	UserConfigFile *config.UserConfigFile

	Component types.Component
	ID        int
}

// Pollster is the interface for datadog metrics polling.
type Pollster interface {
	Do(ctx context.Context) chan *Response
}
