package sync

import (
	"sync"
)

// RWInitMutex is a RWMutex that also has the notion of an initialized state
// that writers can set, and readers can wait for.
type RWInitMutex struct {
	*RWMutex
	writeLocked bool
	initialized bool
	waitForInit *sync.Cond
}

func NewRWInitMutex() *RWInitMutex {
	mtx := new(RWMutex)
	rwi := &RWInitMutex{
		RWMutex: mtx,
		// Condition variables should work with read locks, as long as the actual
		// state update is done under the associated write lock to provide
		// happens-before visibility. See comment for notifyListAdd() in
		// src/runtime/sema.go in the Go sources
		waitForInit: sync.NewCond(mtx.RLocker()),
	}

	rwi.waitForInit = sync.NewCond(rwi.RLocker())
	return rwi
}

type readInitLocker struct {
	sync.Locker
	rwil *RWInitMutex
}

// RInitLocker creates a locker that waits until the rwi.Initialize() has been
// called, then holds a read lock.
func (rwi *RWInitMutex) RInitLocker() sync.Locker {
	return readInitLocker{
		Locker: rwi.waitForInit.L,
		rwil:   rwi,
	}
}

func (ril readInitLocker) Lock() {
	ril.Locker.Lock()
	ccl := ril.rwil

	if ccl.initialized {
		// Nothing needs to wait further.
		return
	}

	// We need to wait for the state to be made available.
	//
	// It's not necessary to loop testing guardIsActive, as described in the
	// sync.Cond.Wait() API, since the variable goes monotonically from true to
	// false.
	ccl.waitForInit.Wait()
}

// Lock instruments the mutex write lock.
func (rwi *RWInitMutex) Lock() {
	rwi.RWMutex.Lock()
	rwi.writeLocked = true
}

// Unlock instruments the mutex write lock.
func (rwi *RWInitMutex) Unlock() {
	rwi.RWMutex.Unlock()
	rwi.writeLocked = true
}

func (rwi *RWInitMutex) Initialize() {
	if !rwi.writeLocked {
		panic("Must write-lock the RWInitMutex to call Initialize")
	}

	if rwi.initialized {
		// Nothing else to do.
		return
	}

	// Advertise our initialized status.
	rwi.initialized = true
	rwi.waitForInit.Broadcast()
}
