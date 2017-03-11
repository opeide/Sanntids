package request_watchdog

/*
* Possibility of timer on each floor.
*/

import(
	"time"
	"fmt"
	"../message_structs"
	"../hardware_interface"
)

const timeout_seconds = 10

//indexed by id and floor. 
var stop_channels = make(map[string][]chan int )
var active_timers = make(map[string][]bool)

var timed_out_requests_chan chan<- message_structs.Request

func Init(timed_out_requests_chan_parameter chan<- message_structs.Request){
	timed_out_requests_chan = timed_out_requests_chan_parameter
}

func Timer_start(request message_structs.Request){
	if _, exists := stop_channels[request.Responsible_elevator]; !exists{
		stop_channels[request.Responsible_elevator] = make([]chan int, hardware_interface.N_FLOORS)
		active_timers[request.Responsible_elevator] = make([]bool, hardware_interface.N_FLOORS)
	}

	if !active_timers[request.Responsible_elevator][request.Floor] {
		active_timers[request.Responsible_elevator][request.Floor] = true
		stop_channels[request.Responsible_elevator][request.Floor] = make(chan int, 1)

		go timer_thread(request)
		fmt.Println("Started timer for: ", request)
	}
}

func Timer_stop(request message_structs.Request){
	if _, exists := stop_channels[request.Responsible_elevator]; exists{
		if active_timers[request.Responsible_elevator][request.Floor]{
			stop_channels[request.Responsible_elevator][request.Floor] <- 1
		}
		fmt.Println("Stopped timer for: ", request)
	}
}

// Meant to run as thread
func timer_thread(request message_structs.Request){
	select{
		case <-stop_channels[request.Responsible_elevator][request.Floor]:
			active_timers[request.Responsible_elevator][request.Floor] = false
		case <-time.After(time.Second * timeout_seconds):
			timed_out_requests_chan <- request
			active_timers[request.Responsible_elevator][request.Floor] = false
	}
}