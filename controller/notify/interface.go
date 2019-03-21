package notify

import (
	"context"
)

// SenderID custom type defines different notification senders
type SenderID int

const (
	notifyGithubComment = SenderID(iota)
	notifySlackChannel
	notifySlackPM
)

// NotificationLevel is a custom type to indicate a message severity.
type NotificationLevel string

const (
	// NSuccess is a success level.
	NSuccess = "SUCCESS"

	// NInfo is information level.
	NInfo = "INFO"

	// NWarning is a warning level.
	NWarning = "WARN"

	// NError is an error level.
	NError = "ERROR"
)

// Sender defines a generic interface for notification services.
// A context object could be used to pass a complex messages as a context value.
type Sender interface {
	ID() SenderID
	Notify(ctx context.Context, level NotificationLevel, title, body string) error
}
