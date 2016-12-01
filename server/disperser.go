package server

import (
	"net"
	"time"

	"sync"

	"github.com/gorilla/websocket"
	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/pkg/errors"
)

// Disperser sends events to external clients through websockets.
// If a client C is a superadmin, C receives all events in the system.
// Else if a client C is an admin for an app A, C receives all events about
// A, including when other admins start listening to A's events. C does not
// receive events when other admins start listening to apps that are not A. C
// does not receive events when superadmins start listening.

type disperser struct {
	listenersLock sync.RWMutex
	listeners     map[string]listener

	events     map[string][]event // keys are app IDs
	recent     []event
	eventsLock sync.RWMutex

	mmdb *maxminddb.Reader
}

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
	OriginalIP    string    `json:"originalIP"`
	OriginalLat   float64   `json:"originalLat"`
	OriginalLong  float64   `json:"originalLong"`
	ResolvingIP   string    `json:"resolvingIP"`
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

func newDisperser(f string) (*disperser, error) {
	mmdb, err := maxminddb.Open(f)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open maxmind DB")
	}

	d := &disperser{
		sync.RWMutex{},
		make(map[string]listener),
		make(eventsMap),
		make([]event, 0, 10000),
		sync.RWMutex{},
		mmdb,
	}

	go d.getMessages()

	return d, nil
}

func (d *disperser) addListener(l listener) {
	d.listenersLock.Lock()

	toSend := map[string]eventName{}
	for short, long := range events {
		toSend[long] = short
	}
	l.conn.WriteJSON(toSend)
	d.listeners[l.conn.LocalAddr().String()] = l
	origHandler := l.conn.CloseHandler()
	l.conn.SetCloseHandler(func(code int, text string) error {
		delete(d.listeners, l.conn.LocalAddr().String())
		return origHandler(code, text)
	})

	d.listenersLock.Unlock()
}

func (d *disperser) addEvent(n eventName, t time.Time, aID, s, uID,
	oIP, rIP string) error {
	if _, found := events[n]; !found {
		return errors.Errorf("%d was not in the event map", n)
	}

	e := event{
		Name:      events[n],
		AppID:     aID,
		Timestamp: t,
		Status:    s,
		UserID:    uID,
	}

	if net.ParseIP(oIP) != nil {
		var oRec struct {
			Location struct {
				AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
				Latitude       float64 `maxminddb:"latitude"`
				Longitude      float64 `maxminddb:"longitude"`
				MetroCode      uint    `maxminddb:"metro_code"`
				TimeZone       string  `maxminddb:"time_zone"`
			} `maxminddb:"location"`
		}
		err := d.mmdb.Lookup(net.ParseIP(oIP), &oRec)
		if err != nil {
			return err
		}

		e.OriginalIP = oIP
		e.OriginalLat = oRec.Location.Latitude
		e.OriginalLong = oRec.Location.Longitude
	}

	if net.ParseIP(rIP) != nil {
		var rRec struct {
			Location struct {
				AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
				Latitude       float64 `maxminddb:"latitude"`
				Longitude      float64 `maxminddb:"longitude"`
				MetroCode      uint    `maxminddb:"metro_code"`
				TimeZone       string  `maxminddb:"time_zone"`
			} `maxminddb:"location"`
		}
		err := d.mmdb.Lookup(net.ParseIP(rIP), &rRec)
		if err != nil {
			return err
		}

		e.ResolvingIP = rIP
		e.ResolvingLat = rRec.Location.Latitude
		e.ResolvingLong = rRec.Location.Longitude
	}

	go func() {
		d.eventsLock.Lock()
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
		d.eventsLock.Unlock()
	}()

	return nil
}

// Every second, aggregates all the events for the last second. Sends the
// results out to the clients.
func (d *disperser) getMessages() {
	defer func() {
		err := d.mmdb.Close()
		if err != nil {
			panic(errors.Wrap(err, "Could not close MaxMind DB"))
		}
	}()

	for true {
		time.Sleep(100 * time.Millisecond)

		// Copy current events
		eventsCopy := make(map[string][]event)
		d.eventsLock.RLock()
		for k, v := range d.events {
			eventsCopy[k] = v
		}
		d.eventsLock.RUnlock()

		// Copy current listeners
		listenersCopy := make(map[string]listener)
		d.listenersLock.RLock()
		for k, v := range d.listeners {
			listenersCopy[k] = v
		}
		d.listenersLock.RUnlock()

		// Send events to listeners
		for _, l := range listenersCopy {
			l.conn.WriteJSON(eventsList{
				Type:   "List",
				Events: eventsCopy[l.appID],
			})
		}
	}
}

func (d *disperser) getRecent() []event {
	var out []event
	d.eventsLock.RLock()
	copy(out, d.recent)
	d.eventsLock.RUnlock()
	return out
}
