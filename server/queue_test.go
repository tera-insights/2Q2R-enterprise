// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"testing"
	"time"
)

var lTimeout = 3 * time.Second
var rcTimeout = 3 * time.Second
var interval = 10 * time.Second
var q = NewQueue(rcTimeout, interval, lTimeout, interval)
var listenerError = "Did not receive proper status listener channel"

func TestListenOnCompleted(t *testing.T) {
	id := randString(32)
	q.MarkCompleted(id)
	c := q.Listen(id)
	if 200 != <-c {
		t.Errorf(listenerError)
	}
}

func TestListenOnLaterCompleted(t *testing.T) {
	id := randString(32)
	c := q.Listen(id)
	q.MarkCompleted(id)
	if 200 != <-c {
		t.Errorf(listenerError)
	}
}

func TestMultipleListeners(t *testing.T) {
	id := randString(32)
	cA := q.Listen(id)
	cB := q.Listen(id)
	q.MarkCompleted(id)
	if 200 != <-cA || 200 != <-cB {
		t.Errorf(listenerError)
	}
}

func TestListenMarkListen(t *testing.T) {
	id := randString(32)
	cA := q.Listen(id)
	q.MarkCompleted(id)
	cB := q.Listen(id)
	if 200 != <-cA || 200 != <-cB {
		t.Errorf(listenerError)
	}
}

func TestListenAndRefuse(t *testing.T) {
	id := randString(32)
	c := q.Listen(id)
	q.MarkRefused(id)
	if 401 != <-c {
		t.Errorf(listenerError)
	}
}

/*func TestListenAndTimeout(t *testing.T) {
	id := randString(32)
	c := q.Listen(id)
	if 408 != <-c {
		t.Errorf(listenerError)
	}
}

// Need to assert that an error is not thrown if the client isn't listening
// when a request times out.
func TestListenAndDropListener(t *testing.T) {
	id := randString(32)
	q.Listen(id)

	// Wait for the timeout
	go func() {
		time.Sleep(lTimeout + 5*time.Second)
	}()
}*/
