package sync_test

// To run the tests: `go test -race ./libs/sync`

import (
	"runtime"
	"testing"
	"time"

	cmtsync "github.com/tendermint/tendermint/libs/sync"
)

const (
	goroutineDuration = 1 * time.Second
)

func workerHelper(mtx *cmtsync.RWInitMutex, opFunc func(*cmtsync.RWInitMutex), done chan<- bool) {
	timer := time.After(goroutineDuration)
	defer func() {
		done <- true
	}()

	for {
		opFunc(mtx)

		select {
		case <-timer:
			return
		default:
			continue
		}
	}
}

func readOperation(mtx *cmtsync.RWInitMutex) {
	mtx.IsInitialized()
}

func writeOperation(mtx *cmtsync.RWInitMutex) {
	mtx.Lock()
	mtx.Initialize()
	defer mtx.Unlock()
}

func rinitOperation(mtx *cmtsync.RWInitMutex) {
	mtx.RInitLock()
	mtx.RInitUnlock()
}

func readWorker(mtx *cmtsync.RWInitMutex, done chan bool) {
	workerHelper(mtx, readOperation, done)
}

func writeWorker(mtx *cmtsync.RWInitMutex, done chan bool) {
	workerHelper(mtx, writeOperation, done)
}

func rinitWorker(mtx *cmtsync.RWInitMutex, done chan bool) {
	workerHelper(mtx, rinitOperation, done)
}

func doTestParallelInitialized(numReaders, gomaxprocs int) {
	const numWriters = 1
	runtime.GOMAXPROCS(gomaxprocs)
	mtx := cmtsync.NewRWInitMutex()
	done := make(chan bool, numReaders+numWriters)

	for i := 0; i < numWriters; i++ {
		go writeWorker(mtx, done)
	}

	for i := 0; i < numReaders; i++ {
		go readWorker(mtx, done)
	}

	for i := 0; i < numWriters; i++ {
		<-done
	}

	for i := 0; i < numReaders; i++ {
		<-done
	}
}

func TestInitialized(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(-1))
	doTestParallelInitialized(2, 4)
	doTestParallelInitialized(3, 4)
	doTestParallelInitialized(4, 4)
}

func doTestRLock(gomaxprocs int) {
	const workers = 3
	runtime.GOMAXPROCS(gomaxprocs)
	mtx := cmtsync.NewRWInitMutex()
	done := make(chan bool, workers)

	go rinitWorker(mtx, done)
	go writeWorker(mtx, done)
	go rinitWorker(mtx, done)

	for i := 0; i < workers; i++ {
		<-done
	}
}

func TestRLock(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(-1))
	doTestRLock(2)
	doTestRLock(3)
	doTestRLock(4)
}

func doTestInterleave(numReaders, gomaxprocs int) {
	const workers = 3
	runtime.GOMAXPROCS(gomaxprocs)
	mtx := cmtsync.NewRWInitMutex()
	done := make(chan bool, workers)

	for i := 0; i < numReaders; i++ {
		go readWorker(mtx, done)
	}

	go rinitWorker(mtx, done)
	go writeWorker(mtx, done)
	go rinitWorker(mtx, done)

	for i := 0; i < numReaders; i++ {
		<-done
	}

	for i := 0; i < workers; i++ {
		<-done
	}
}

func TestInterleave(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(-1))
	doTestInterleave(2, 4)
	doTestInterleave(2, 5)
	doTestInterleave(2, 6)
}
