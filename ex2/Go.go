// Go 1.2
// go run helloworld_go.go

package main

import (
	"fmt"
	"runtime"
	"time"
)

func thread1_func(ch chan int) {
	for j := 0; j < 1000000; j++ {
		i := <-ch
		i++
		ch <- i
	}
}

func thread2_func(ch chan int) {
	for j := 0; j < 1000000; j++ {
		i := <-ch
		i--
		ch <- i
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // I guess this is a hint to what GOMAXPROCS does...

	// Try doing the exercise both with and without it!

	ch := make(chan int, 1)
	ch <- 0

	go thread1_func(ch)
	go thread2_func(ch)

	// We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
	// We'll come back to using channels in Exercise 2. For now: Sleep.
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("num: %v\n", <-ch)
}
