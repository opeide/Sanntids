package request_executor

import (
	"../elev"
	"../request"
)

var last_motor_direction int
var last_visited_floor int

type elevator_state_t int

var elevator_state elevator_state_t

const (
	starting_up elevator_state = iota
	at_floor_idle
	at_floor_doors_open
	moving_down
	moving_up
)

var requests_upward [elev.N_FLOORS]request.Request
var requests_downward [elev.N_FLOORS]request.Requests

func Execute_requests(requests_to_execute_chan <-chan request.Request, executed_requests_chan chan<- requesst.Request) {
	for {
		select {
		case request_to_execute := <-requests_to_execute_chan:
			switch request_to_execute.Direction {
				case elev.MOTOR_DIRECTION DOWN:
					if requests_downward[request_to_execute.Floor]

			}
		}
	}
}
