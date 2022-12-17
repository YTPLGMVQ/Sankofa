package main

import (
	"sankofa/ow"
	"sync"
)

////////////////////////////////////////////////////////////////
// A swarm of workers.
////////////////////////////////////////////////////////////////

// number of go routines should be large enough to keep
// all the cores of a modern processor busy
var goroutines int

// synchronization
var waitGroup sync.WaitGroup

// concurrent access to several variables
var mutex sync.RWMutex

// count changed ranks
var cnt int

func worker(v func(r int64), input <-chan int64) {
	for rank := range input {
		v(rank)
	}
	waitGroup.Done()
}

func incCounter() {
	mutex.Lock()
	defer mutex.Unlock()
	cnt += 1
	ow.Log("counter:", cnt)
}

func counter() int {
	mutex.RLock()
	defer mutex.RUnlock()
	ow.Log("counter:", cnt)
	return cnt
}

func startWorkers(v func(r int64)) chan<- int64 {
	ow.Log("starting", goroutines, "workers")
	channel := make(chan int64, goroutines)
	cnt = 0
	for i := 0; i < goroutines; i++ {
		waitGroup.Add(1)
		go worker(v, channel)
	}
	return channel
}

func stopWorkers(output chan<- int64) {
	ow.Log("stopping workers")
	close(output)
	waitGroup.Wait()
	ow.Log("...stopped")
}
