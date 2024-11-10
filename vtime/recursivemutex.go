// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import "sync"

type recursiveRWLock struct {
	rw      sync.RWMutex // used for locking for writes
	rMutex  sync.Mutex   // protects the reader count
	readers int          // keeps track of the number of active readers
}

// RLock acquires the read lock recursively.
func (rrw *recursiveRWLock) RLock() {
	rrw.rMutex.Lock()
	rrw.readers++
	if rrw.readers == 1 {
		rrw.rw.RLock()
	}
	rrw.rMutex.Unlock()
}

// RUnlock releases the read lock.
func (rrw *recursiveRWLock) RUnlock() {
	rrw.rMutex.Lock()
	rrw.readers--
	if rrw.readers == 0 {
		rrw.rw.RUnlock()
	}
	rrw.rMutex.Unlock()
}

// Lock acquires the write lock.
func (rrw *recursiveRWLock) Lock() {
	rrw.rw.Lock()
}

// Unlock releases the write lock.
func (rrw *recursiveRWLock) Unlock() {
	rrw.rw.Unlock()
}
