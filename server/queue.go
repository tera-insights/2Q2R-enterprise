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
// 2. If not, add it to a list of listeners for thar request
//
// When a new request completion comes in:
// 1. Alert all listeners that the request was completed
// 2. Add it to the recently completed list
//
// Cleans out the both the recently completed list and waiting lists
// at fixed time intervals.

/**

QUESTIONS:
1. If a listener is evicted, we send 408 to it. However, if that channel is
   not listening do we die?
2. When a client calls MarkCompleted, we send 401 to all the listeners in that
   queue. We need to assert that we FIRST add the requestID to the recently
   completed list. This is because a new Listen request might come in while
   we're iterating over the listeners. This request would be added to the NEW
   listeners set, hence not iterated over. Moreover it would never check
   recentlyCompleted.
*/
type Queue struct {
	recentlyCompleted *cache.Cache
	rcTimeout         time.Duration
	rcInterval        time.Duration
	listeners         *cache.Cache
	lTimeout          time.Duration
	lInterval         time.Duration
}

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
			listener <- http.StatusRequestTimeout
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
	if success, found := q.recentlyCompleted.Get(requestID); found {
		if success.(bool) {
			c <- http.StatusOK
		} else {
			c <- http.StatusUnauthorized
		}
		return c
	}
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		newListeners := append(listeners, c)
		q.listeners.Set(requestID, newListeners, q.lTimeout)
	} else {
		q.listeners.Set(requestID, []chan int{c}, q.lTimeout)
	}
	return c
}

// MarkCompleted records that a request was successfully completed.
func (q Queue) MarkCompleted(requestID string) {
	q.recentlyCompleted.Set(requestID, true, q.rcTimeout)
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		for _, element := range listeners {
			element <- http.StatusOK
		}
		q.listeners.Delete(requestID)
	}
}

// MarkRefused records that a request was refused.
func (q Queue) MarkRefused(requestID string) {
	q.recentlyCompleted.Set(requestID, false, q.rcTimeout)
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan int)
		for _, element := range listeners {
			element <- http.StatusUnauthorized
		}
		q.listeners.Delete(requestID)
	}
}
