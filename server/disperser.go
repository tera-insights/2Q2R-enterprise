package server

import (
	"time"

	"github.com/gorilla/websocket"
)

// Receiver keeps track of events (receives them from the outside world)
type disperser struct {
	eventInput      chan string
	aggregatorInput chan chan []string
	events          []string
	listeners       []*websocket.Conn
}

func newDisperser() *disperser {
	return &disperser{
		make(chan string, 10000),
		make(chan chan []string),
		[]string{},
		[]*websocket.Conn{},
	}
}

func (d *disperser) addListener(c *websocket.Conn) {
	d.listeners = append(d.listeners, c)
}

func (d *disperser) addEvent(s string) {
	d.eventInput <- s
}

// Either does nothing, adds a new event to the current list of events, or
// sends the slice of events.
func (d *disperser) listen() {
	for true {
		select {
		case output := <-d.aggregatorInput:
			old := d.events
			d.events = []string{}
			output <- old
		case event := <-d.eventInput:
			d.events = append(d.events, event)
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
		eventsChan := make(chan []string)
		d.aggregatorInput <- eventsChan
		events := <-eventsChan
		for _, conn := range d.listeners {
			conn.WriteJSON(eventSummary{
				Type:   "Summary",
				Events: events,
			})
			conn.WriteJSON(events)
		}
	}
}
