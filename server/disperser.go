package server

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type listener struct {
	conn  *websocket.Conn
	appID string // if 1, receives all events
}

type eventName int

const (
	listenerRegistered eventName = iota
	login
	registration
	keyDeletion
)

var events = map[eventName]string{
	listenerRegistered: "listenerRegistered",
	login:              "login",
	registration:       "registration",
	keyDeletion:        "keyDeletion",
}

type event struct {
	Name  string `json:"name"`
	AppID string `json:"appID"`
}

type eventsMap map[string][]event

type eventSummary struct {
	Type   string  `json:"__type__"` // type of message
	Events []event `json:"events"`
}

// Receiver keeps track of events (receives them from the outside world)
type disperser struct {
	eventInput chan event

	// send chans that should receive the event map
	aggregatorOutput chan chan eventsMap

	// send chans that should receive the most recent events
	recentOutput chan chan []event

	events    map[string][]event // keys are app IDs
	listeners []listener
	recent    []event
}

func newDisperser() *disperser {
	return &disperser{
		make(chan event, 10000),
		make(chan chan eventsMap),
		make(chan chan []event),
		make(eventsMap),
		[]listener{},
		make([]event, 0, 10000),
	}
}

func (d *disperser) addListener(l listener) {
	toSend := map[string]eventName{}
	for short, long := range events {
		toSend[long] = short
	}
	l.conn.WriteJSON(toSend)
	d.listeners = append(d.listeners, l)
}

func (d *disperser) addEvent(n eventName, ids []string) error {
	if _, found := events[n]; !found {
		return errors.Errorf("%d was not in the event map", n)
	}
	sentToGlobal := false
	for _, id := range ids {
		if id == "1" {
			sentToGlobal = true
		}
		d.eventInput <- event{events[n], id}
	}
	// Add event to the global events list if it not done above
	if !sentToGlobal {
		d.eventInput <- event{events[n], "1"}
	}
	return nil
}

// Either does nothing, adds a new event to the current list of events, or
// sends the slice of events.
func (d *disperser) listen() {
	for true {
		select {
		case where := <-d.aggregatorOutput:
			old := d.events
			d.events = make(map[string][]event)
			where <- old
		case where := <-d.recentOutput:
			where <- d.recent
		case e := <-d.eventInput:
			d.events[e.AppID] = append(d.events[e.AppID], e)
			// If the recent list is full, overwrite the oldest event
			if len(d.recent) == cap(d.recent) {
				copy(d.recent, append(d.recent[1:], e))
			} else {
				d.recent = append(d.recent, e)
			}
		default:
			// No message! do nothing
		}
	}
}

// Every second, aggregates all the events for the last second. Sends the
// results out to the clients.
func (d *disperser) getMessages() {
	for true {
		time.Sleep(time.Second)
		ec := make(chan eventsMap)
		d.aggregatorOutput <- ec
		events := <-ec
		for _, l := range d.listeners {
			l.conn.WriteJSON(eventSummary{
				Type:   "Summary",
				Events: events[l.appID],
			})
		}
	}
}

func (d *disperser) getRecent() []event {
	out := make(chan []event, 1)
	d.recentOutput <- out
	return <-out
}
