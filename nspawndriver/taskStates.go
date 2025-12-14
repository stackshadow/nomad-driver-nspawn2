// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package nspawndriver

import (
	"sync"
)

// taskState provides a mechanism to store and retrieve
// task handles given a string identifier. The ID should
// be unique per task
type taskStates struct {
	store map[string]*taskState
	lock  sync.RWMutex
}

func newTaskStore() *taskStates {
	return &taskStates{store: map[string]*taskState{}}
}

func (ts *taskStates) Set(id string, handle *taskState) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	ts.store[id] = handle
}

func (ts *taskStates) Get(id string) (*taskState, bool) {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	t, ok := ts.store[id]
	return t, ok
}

func (ts *taskStates) Delete(id string) {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	delete(ts.store, id)
}
