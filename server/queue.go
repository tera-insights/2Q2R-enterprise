// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
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
type Queue struct {
	recentlyCompleted *cache.Cache // Maps request ID to status code
	rcTimeout         time.Duration
	rcInterval        time.Duration
	listeners         *cache.Cache
	lTimeout          time.Duration
	lInterval         time.Duration
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

// NewQueue initializes a new queue.
func NewQueue(rcTimeout time.Duration, rcInterval time.Duration,
	lTimeout time.Duration, lInterval time.Duration) Queue {
	q := Queue{
		cache.New(rcTimeout, rcInterval),
		rcTimeout,
		rcInterval,
		cache.New(lTimeout, lInterval),
		lTimeout,
		lInterval,
	}
	q.listeners.OnEvicted(func(requestID string, data interface{}) {
		listeners := data.([]chan int)
		for _, listener := range listeners {
			signal(listener, http.StatusRequestTimeout)
			close(listener)
		}
	})
	return q
}

// Listen returns a channel that will emit the status code for the request. The
// options are:
// 200, for success
// 401, for requests canceled by the user
// 408, for requests that timed out
func (q Queue) Listen(requestID string) chan int {
	c := make(chan int, 1)
	if status, found := q.recentlyCompleted.Get(requestID); found {
		signal(c, status.(int))
		return c
	}
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		newListeners := append(listeners, c)
		q.listeners.Set(requestID, newListeners, q.lTimeout)
	} else {
		q.listeners.Set(requestID, []chan int{c}, q.lTimeout)
		go func() {
			time.Sleep(q.lTimeout)
			q.recentlyCompleted.Set(requestID, http.StatusRequestTimeout,
				q.rcTimeout)
			q.listeners.Delete(requestID)
		}()
	}
	return c
}

// MarkCompleted records that a request was successfully completed.
func (q Queue) MarkCompleted(requestID string) {
	q.recentlyCompleted.Set(requestID, http.StatusOK, q.rcTimeout)
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		for _, listener := range listeners {
			signal(listener, http.StatusOK)
		}
		q.listeners.Delete(requestID)
	}
}

// MarkRefused records that a request was refused.
func (q Queue) MarkRefused(requestID string) {
	q.recentlyCompleted.Set(requestID, http.StatusUnauthorized, q.rcTimeout)
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		for _, listener := range listeners {
			signal(listener, http.StatusUnauthorized)
		}
		q.listeners.Delete(requestID)
	}
}
