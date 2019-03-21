package notify

import (
	"context"
	"strings"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	rate = time.Second
)

const (
	priorityLow    = 1
	priorityMedium = 3
	priorityHigh   = 5
)

// NewSlackSender returns a new instance of slack notification service which implements Sender interface.
func NewSlackSender(ctx context.Context, apiToken string) Sender {
	slackNotifier := &slackNotification{
		api: slack.New(apiToken),
		q:   queue.NewPriorityQueue(500),
	}

	go slackNotifier.watchItems(ctx, rate)

	return slackNotifier
}

type slackNotification struct {
	api     *slack.Client
	channel string

	q *queue.PriorityQueue
}

func (s slackNotification) watchItems(ctx context.Context, waitTime time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
			if s.q.Empty() {
				continue
			}

			if err := s.notify(ctx); err != nil {
				logrus.Errorf("Error notifying using slack backend: %s", err)
			}
		}
	}
}

// SenderID returns a unique sender's ID.
func (s slackNotification) ID() SenderID {
	return notifySlackChannel
}

// Notify sends a notification to slack.
func (s slackNotification) Notify(ctx context.Context, level NotificationLevel, title, body string) error {
	if s.channel == "" || title == "" {
		return errors.New("slack channel and title are required")
	}

	color := "good"
	priority := priorityLow
	switch level {
	case NInfo:
		color = "#439FE0"
	case NError:
		color = "danger"
		priority = priorityHigh
	case NWarning:
		color = "warning"
		priority = priorityMedium
	}

	return s.q.Put(&Message{
		ctx:      ctx,
		priority: priority,
		color:    color,

		Title:   title,
		Body:    body,
		Channel: s.channel,
	})
}

func (s slackNotification) notify(ctx context.Context) error {
	messages, err := s.q.Get(1)
	if err != nil {
		return err
	}

	var errs []string
	for _, msg := range messages {
		message, ok := msg.(*Message)
		if !ok {
			return errors.New("unable to type assert to Message")
		}

		// wrap the contexts
		ctx, cancel := context.WithCancel(ctx)
		ctx, cancel = context.WithCancel(message.ctx)

		_, _, _, err = s.api.SendMessageContext(ctx, message.Channel, slack.MsgOptionAttachments(message.attachment()))
		if err != nil {
			errs = append(errs, err.Error())
		}
		cancel()
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Errorf("the following errors were while sending slack notification: %s", strings.Join(errs, "; "))
}

// Message is an object of a slack message
type Message struct {
	ctx      context.Context
	priority int // priority of the item in the queue.
	color    string

	Title   string
	Body    string
	Channel string
}

func (m Message) attachment() slack.Attachment {
	return slack.Attachment{
		Color:   m.color,
		Pretext: m.Title,
		Text:    m.Body,
	}
}

// Compare is an implementation of Item object used by a queue.
func (m Message) Compare(other queue.Item) int {
	// the items with greater priority value have more priority
	otherMsg := other.(*Message)
	if otherMsg.priority > m.priority {
		return 1
	} else if otherMsg.priority < m.priority {
		return -1
	}
	return 0
}
