// Go 1.2
// go run helloworld_go.go

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var i int = 0

func thread1_func(mutex *sync.Mutex) {
	for j := 0; j < 1000000; j++ {
		mutex.Lock()
		i++
		mutex.Unlock()
	}
}

func thread2_func(mutex *sync.Mutex) {
	for j := 0; j < 1000000; j++ {
		mutex.Lock()
		i--
		mutex.Unlock()
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // I guess this is a hint to what GOMAXPROCS does...

	// Try doing the exercise both with and without it!

	var mutex = &sync.Mutex{}

	go thread1_func(mutex)
	go thread2_func(mutex)

	// We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
	// We'll come back to using channels in Exercise 2. For now: Sleep.
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("num: %v\n", i)
}
