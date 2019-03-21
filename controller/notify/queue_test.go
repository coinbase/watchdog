package notify

import (
	"github.com/golang-collections/go-datastructures/queue"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	q := queue.NewPriorityQueue(100)

	msgs := []queue.Item{
		&Message{
			Title:    "foo",
			priority: 5,
		},
		&Message{
			Title:    "foo2",
			priority: 10,
		},
		&Message{
			Title:    "foo3",
			priority: 1,
		},
		&Message{
			Title:    "foo3",
			priority: 1,
		},
	}

	if q.Len() != 0 {
		t.Fatalf("queue is not empty: %d", q.Len())
	}

	err := q.Put(msgs...)
	if err != nil {
		t.Fatal(err)
	}

	if q.Len() != 3 {
		t.Fatalf("queue must have 3 items. Got %d", q.Len())
	}

	items, err := q.Get(3)
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 3 {
		t.Fatalf("queue must return 3 items. Got %d", len(items))
	}

	item1 := toMessage(items[0])
	item2 := toMessage(items[1])
	item3 := toMessage(items[2])

	if item1.Title != "foo2" {
		t.Fatalf("expect item with title foo2. Got %+v", item1)
	}

	if item2.Title != "foo" {
		t.Fatalf("expect item with title foo. Got %+v", item2)
	}

	if item3.Title != "foo3" {
		t.Fatalf("expect item with name foo3. Got %+v", item3)
	}
}

func toMessage(item queue.Item) *Message {
	return item.(*Message)
}
