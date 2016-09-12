// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"testing"
	"time"
)

var lTimeout = 1 * time.Minute
var rcTimeout = 3 * time.Second
var interval = 10 * time.Second
var q = NewQueue(rcTimeout, interval, lTimeout, interval)
var listenerError = "Did not receive success message from listener channel"

func TestListenOnCompleted(t *testing.T) {
	id := "foo"
	q.MarkCompleted(id)
	c := q.Listen(id)
	if true != <-c {
		t.Errorf(listenerError)
	}
}

func TestListenOnLaterCompleted(t *testing.T) {
	id := "foo"
	c := q.Listen(id)
	q.MarkCompleted(id)
	if true != <-c {
		t.Errorf(listenerError)
	}
}

func TestMultipleListeners(t *testing.T) {
	id := "foo"
	cA := q.Listen(id)
	cB := q.Listen(id)
	q.MarkCompleted(id)
	if true != <-cA || true != <-cB {
		t.Errorf(listenerError)
	}
}

func TestListenMarkListen(t *testing.T) {
	id := "foo"
	cA := q.Listen(id)
	q.MarkCompleted(id)
	cB := q.Listen(id)
	if true != <-cA || true != <-cB {
		t.Errorf(listenerError)
	}
}
