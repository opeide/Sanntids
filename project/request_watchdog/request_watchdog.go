package request_watchdog

import (
	"../hardware_interface"
	"../message_structs"
	"os"
	"time"
)

const timeout_seconds = 40 //time to move up and down stopping everywhere

//indexed by id and floor.
var stop_channels = make(map[string][]chan int)
var active_timers = make(map[string][]bool)

var timed_out_requests_chan chan<- message_structs.Request
var local_id string

func Init(local_id_parameter string, timed_out_requests_chan_parameter chan<- message_structs.Request) {
	timed_out_requests_chan = timed_out_requests_chan_parameter
	local_id = local_id_parameter
}

func Timer_start(request message_structs.Request) {
	if _, exists := stop_channels[request.Responsible_elevator]; !exists {
		stop_channels[request.Responsible_elevator] = make([]chan int, hardware_interface.N_FLOORS)
		active_timers[request.Responsible_elevator] = make([]bool, hardware_interface.N_FLOORS)
	}

	if !active_timers[request.Responsible_elevator][request.Floor] {
		active_timers[request.Responsible_elevator][request.Floor] = true
		stop_channels[request.Responsible_elevator][request.Floor] = make(chan int, 1)

		go timer_thread(request)
	}
}

func Timer_stop(request message_structs.Request) {
	if _, exists := stop_channels[request.Responsible_elevator]; exists {
		if active_timers[request.Responsible_elevator][request.Floor] {
			stop_channels[request.Responsible_elevator][request.Floor] <- 1
		}
	}
}

// Meant to run as thread
func timer_thread(request message_structs.Request) {
	select {
	case <-stop_channels[request.Responsible_elevator][request.Floor]:
		active_timers[request.Responsible_elevator][request.Floor] = false
	case <-time.After(time.Second * timeout_seconds):
		if request.Responsible_elevator == local_id {
			os.Exit(0) //Lets backup take over (effectively a program restart)
		}
		timed_out_requests_chan <- request
		active_timers[request.Responsible_elevator][request.Floor] = false
	}
}
