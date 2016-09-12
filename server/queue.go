// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

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
