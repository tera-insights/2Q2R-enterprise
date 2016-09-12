// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
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
type Queue struct {
	listeners                *cache.Cache
	recentlyCompleted        *cache.Cache
	recentlyCompletedTimeout time.Duration
	listenersTimeout         time.Duration
}

// Listen returns true, nil if the request was already completed and we have it
// in the cache, then returns (true, nil). Else returns false, r where r is a
// pointer to a chan. r will send true if and when the request completes and
// will send false if the request listeners time out before the request
// completes.
func (q Queue) Listen(requestID string) chan bool {
	c := make(chan bool, 1)
	if _, found := q.recentlyCompleted.Get(requestID); found {
		c <- true
		return c
	}
	if cached, found := q.listeners.Get(requestID); found {
		listeners := cached.([]chan bool)
		newListeners := append(listeners, c)
		q.listeners.Set(requestID, newListeners, q.listenersTimeout)
	} else {
		q.listeners.Set(requestID, []chan bool{c}, q.listenersTimeout)
	}
	return c
}

// MarkCompleted records that a request was completed.
func (q Queue) MarkCompleted(requestID string) {
	q.recentlyCompleted.Set(requestID, true, q.recentlyCompletedTimeout)
	if cached, found := q.listeners.Get(requestID); found {
		var listeners = cached.([]chan bool)
		for _, element := range listeners {
			element <- true
		}
		q.listeners.Delete(requestID)
	}
}
