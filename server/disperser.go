package server

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Disperser sends events to external clients through websockets.
// If a client C is a superadmin, C receives all events in the system.
// Else if a client C is an admin for an app A, C receives all events about
// A, including when other admins start listening to A's events. C does not
// receive events when other admins start listening to apps that are not A. C
// does not receive events when superadmins start listening.

type listener struct {
	conn  *websocket.Conn
	appID string // if 1, receives all events
}

type eventName int

const (
	listenerRegistered eventName = iota
	authentication
	registration
	keyDeletion
)

var events = map[eventName]string{
	listenerRegistered: "listenerRegistered",
	authentication:     "authentication",
	registration:       "registration",
	keyDeletion:        "keyDeletion",
}

type event struct {
	Name          string    `json:"name"`
	OriginalIP    net.IP    `json:"originalIP"`
	OriginalLat   float64   `json:"originalLat"`
	OriginalLong  float64   `json:"originalLong"`
	ResolvingIP   net.IP    `json:"resolvingIP"`
	ResolvingLat  float64   `json:"resolvingLat"`
	ResolvingLong float64   `json:"resolvingLong"`
	AppID         string    `json:"appID"`
	Timestamp     time.Time `json:"when"`
	Status        string    `json:"status"` // success, failure, or timeout
	UserID        string    `json:"userID"`
}

type eventsMap map[string][]event

// Lists the events that happened over the previous 100 ms
type eventsList struct {
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

func (d *disperser) addEvent(n eventName, t time.Time, aID, s, uID string,
	oIP net.IP, rIP net.IP) error {
	if _, found := events[n]; !found {
		return errors.Errorf("%d was not in the event map", n)
	}
	d.eventInput <- event{
		Name:        events[n],
		AppID:       aID,
		Timestamp:   t,
		Status:      s,
		UserID:      uID,
		OriginalIP:  oIP,
		ResolvingIP: rIP,
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

			// Add event to the global events list if it not done above
			if e.AppID != "1" {
				d.events["1"] = append(d.events[e.AppID], e)
			}

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
		time.Sleep(100 * time.Millisecond)
		ec := make(chan eventsMap)
		d.aggregatorOutput <- ec
		events := <-ec
		for _, l := range d.listeners {
			l.conn.WriteJSON(eventsList{
				Type:   "List",
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
