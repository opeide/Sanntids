package motor_movement_watchdog

import (
	"fmt"
	"os"
	"time"
)

const timeout_seconds = 7

var stop_channel = make(chan int)
var timer_has_stopped_chan = make(chan int)
var timer_is_active = false

func Timer_start() {
	if !timer_is_active {
		timer_is_active = true
		go timer_thread()
	}
}

func Timer_stop() {
	if timer_is_active {
		stop_channel <- 1
	}
	select {
	case <-timer_has_stopped_chan:
	}
}

// Meant to run as thread
func timer_thread() {
	select {
	case <-stop_channel:
		timer_is_active = false
		timer_has_stopped_chan <- 1
	case <-time.After(time.Second * timeout_seconds):
		fmt.Println("MOTOR MOVEMENT WATCHDOG TIMED OUT. Restarting...")
		os.Exit(0) //Lets backup take over (effectively a program restart)
	}
}
