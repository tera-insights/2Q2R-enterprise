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

	events      map[string][]event // keys are app IDs
	listeners   []listener
	recent      [10000]event
	recentIndex int
}

func newDisperser() *disperser {
	return &disperser{
		make(chan event, 10000),
		make(chan chan eventsMap),
		make(chan chan []event),
		make(eventsMap),
		[]listener{},
		[10000]event{},
		0,
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

func (d *disperser) addEvent(n eventName, id string) error {
	if _, found := events[n]; !found {
		return errors.Errorf("%d was not in the event map", n)
	}
	d.eventInput <- event{events[n], id}
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
			recent := make([]event, len(d.recent))
			for i := 0; i < len(d.recent); i++ {
				recent[i] = d.recent[(d.recentIndex+i)%len(d.recent)]
			}
			where <- recent
		case e := <-d.eventInput:
			d.events[e.AppID] = append(d.events[e.AppID], e)
			// Add event to the global events list if it not done above
			if e.AppID != "1" {
				d.events["1"] = append(d.events["1"], e)
			}
			d.recent[d.recentIndex] = e
			d.recentIndex = (d.recentIndex + 1) % len(d.recent)
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
