package server

import (
	"time"

	"github.com/gorilla/websocket"
)

type listener struct {
	conn  *websocket.Conn
	appID string // if 1, receives all events
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
	eventInput      chan event
	aggregatorInput chan chan eventsMap
	events          map[string][]event // keys are app IDs
	listeners       []listener
}

func newDisperser() *disperser {
	return &disperser{
		make(chan event, 10000),
		make(chan chan eventsMap),
		make(eventsMap),
		[]listener{},
	}
}

func (d *disperser) addListener(l listener) {
	d.listeners = append(d.listeners, l)
}

func (d *disperser) addEvent(e event) {
	d.eventInput <- e
}

// Either does nothing, adds a new event to the current list of events, or
// sends the slice of events.
func (d *disperser) listen() {
	for true {
		select {
		case output := <-d.aggregatorInput:
			old := d.events
			d.events = make(map[string][]event)
			output <- old
		case e := <-d.eventInput:
			d.events[e.AppID] = append(d.events[e.AppID], e)
			// Add event to the global events list if it not done above
			if e.AppID != "1" {
				d.events["1"] = append(d.events["1"], e)
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
		eventsChan := make(chan eventsMap)
		d.aggregatorInput <- eventsChan
		events := <-eventsChan
		for _, l := range d.listeners {
			l.conn.WriteJSON(eventSummary{
				Type:   "Summary",
				Events: events[l.appID],
			})
		}
	}
}
