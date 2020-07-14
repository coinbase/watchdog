package datadog

import (
	"encoding/json"

	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/types"

	"github.com/pkg/errors"
)

var (
	// ErrNilFunction is returned if the passed parameter is nil.
	ErrNilFunction = errors.New("nil argument function")

	// ErrInvalidFunctionType is returned if the function does not have the correct signature.
	ErrInvalidFunctionType = errors.New("invalid function type")
)

// Option is a functional parameter interface for datadog constructor
type Option func(datadog *Datadog) error

// WithAccessorGetFn is a functional parameter to set the get functions.
func WithAccessorGetFn(component types.Component, fn func(string) (json.RawMessage, error)) Option {
	return func(dd *Datadog) error {
		if fn == nil {
			return ErrNilFunction
		}

		switch component {
		case types.ComponentDashboard:
			dd.getDashboardFn = fn
		case types.ComponentDowntime:
			dd.getDowntimeFn = fn
		case types.ComponentScreenboard:
			dd.getScreenBoardFn = fn
		default:
			return ErrInvalidFunctionType
		}

		return nil
	}
}

// WithMonitorGetFn is a functional parameter to set the dashboard get function.
func WithMonitorGetFn(fn func(string, bool) (*client.MonitorWithDependencies, error)) Option {
	return func(dd *Datadog) error {
		if fn == nil {
			return ErrNilFunction
		}

		dd.getMonitorFullFn = fn

		return nil
	}
}

// WithAccessorUpdateFn is a functional parameter to set the update functions.
func WithAccessorUpdateFn(component types.Component, fn func(json.RawMessage) error) Option {
	return func(dd *Datadog) error {
		if fn == nil {
			return ErrNilFunction
		}

		switch component {
		case types.ComponentDashboard:
			dd.updateDashboardFn = fn
		case types.ComponentDowntime:
			dd.updateDowntimeFn = fn
		case types.ComponentScreenboard:
			dd.updateScreenBoardFn = fn
		default:
			return ErrInvalidFunctionType
		}

		return nil
	}
}

// WithMonitorSetFn is a functional parameter that sets the update monitor function.
func WithMonitorSetFn(fn func(*client.MonitorWithDependencies) error) Option {
	return func(dd *Datadog) error {
		if fn == nil {
			return ErrNilFunction
		}

		dd.updateMonitorFn = fn
		return nil
	}
}
