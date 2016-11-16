// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// Queue lets clients know when a certain request has been fulfilled.
// When a new listener comes in:
// 1. Check if the request was "recently completed"
// 2. If not, add it to a list of listeners for that request
// 3. After listenersTimeout, we need to alert this list of listeners that
//    the request has timed out. We can't rely on go-cache to do this because
//    that timeout resets everytime we call set.
//    We also want to make sure that we don't send more than one message to the
//    channel because it's buffered with one slot.
// 4. Could set up a goroutine to, after lTimeout seconds, send the timeout
//    message to all the listeners for that requestID.
//
// When a new request completion comes in:
// 1. Alert all listeners that the request was completed
// 2. Add it to the recently completed list
//
// Cleans out the both the recently completed list and waiting lists
// at fixed time intervals. Note that this means that some requests will wait
// for the full (e.g.) 3 minutes, and others will wait for less.
type queue struct {
	recentlyCompleted *cache.Cache // Maps request ID to status code
	rcTimeout         time.Duration
	rcInterval        time.Duration
	listeners         *cache.Cache
	lTimeout          time.Duration
	lInterval         time.Duration

	newListeners chan newListener
	completed    chan string
	timedOut     chan string
}

type newListener struct {
	id        string
	out       chan newListenerResponse
	exclusive bool
}

type newListenerResponse struct {
	err error
	out chan int
}

// non-blocking send.
func signal(c chan int, code int) {
	select {
	case c <- code:
		return
	default:
		return
	}
}

// Initializes a new queue.
func newQueue(rcTimeout time.Duration, rcInterval time.Duration,
	lTimeout time.Duration, lInterval time.Duration) queue {
	q := queue{
		cache.New(rcTimeout, rcInterval),
		rcTimeout,
		rcInterval,
		cache.New(lTimeout, lInterval),
		lTimeout,
		lInterval,
		make(chan newListener),
		make(chan string),
		make(chan string),
	}
	q.listeners.OnEvicted(func(requestID string, data interface{}) {
		listeners := data.([]chan int)
		for _, listener := range listeners {
			signal(listener, http.StatusRequestTimeout)
			close(listener)
		}
	})
	go q.listen()
	return q
}

// Listen returns a channel that will emit the status code for the request. The
// options are:
// 200, for success
// 401, for requests canceled by the user
// 408, for requests that timed out
func (q queue) Listen(id string, exclusive bool) (chan int, error) {
	c := make(chan newListenerResponse)
	q.newListeners <- newListener{id, c, exclusive}
	r := <-c
	return r.out, r.err
}

// MarkCompleted records that a request was successfully completed.
func (q queue) MarkCompleted(requestID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			q.recentlyCompleted.Delete(requestID)
			err = errors.Errorf("Could not mark request as completed. "+
				"Panicked with value %s", r)
		}
	}()

	q.completed <- requestID
	return
}

func (q *queue) listen() {
	for {
		select {
		case id := <-q.completed:
			q.recentlyCompleted.Set(id, http.StatusOK, q.rcTimeout)
			if cached, found := q.listeners.Get(id); found {
				listeners := cached.([]chan int)
				for _, listener := range listeners {
					signal(listener, http.StatusOK)
				}
				q.listeners.Delete(id)
			}
		case id := <-q.timedOut:
			q.recentlyCompleted.Set(id, http.StatusRequestTimeout, q.rcTimeout)
			q.listeners.Delete(id)
		case nl := <-q.newListeners:
			var err error
			c := make(chan int, 1)
			if status, found := q.recentlyCompleted.Get(nl.id); found {
				signal(c, status.(int))
				nl.out <- newListenerResponse{err, c}
			}
			if val, found := q.listeners.Get(nl.id); found {
				if nl.exclusive {
					err = errors.New("Someone is already listening")
				} else {
					listeners := val.([]chan int)
					newListeners := append(listeners, c)
					q.listeners.Set(nl.id, newListeners, q.lTimeout)
				}
			} else {
				q.listeners.Set(nl.id, []chan int{c}, q.lTimeout)
				go func() {
					time.Sleep(q.lTimeout)
					q.timedOut <- nl.id
				}()
			}
			nl.out <- newListenerResponse{err, c}
		default:
			// No message! Do nothing
		}
	}
}
