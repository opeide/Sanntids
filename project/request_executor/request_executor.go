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
var requests_downward [elev.N_FLOORS]request.Request

func Execute_requests(requests_to_execute_chan <-chan request.Request, executed_requests_chan chan<- request.Request, floor_changes_chan <-chan int) {
	for {
		select {
		case request_to_execute := <-requests_to_execute_chan:
			switch request_to_execute.Direction {
				case elev.BUTTON_TYPE_CALL_DOWN:
					if len(requests_downward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}
				case elev.BUTTON_TYPE_CALL_UP:
					if len(requests_upward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}	
				case elev.BUTTON_TYPE_COMMAND:
					if len(requests_downward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}
					if len(requests_upward[request_to_execute.Floor]) == 0{
						requests_upward[request_to_execute.Floor] = request_to_execute
					}	
			}
		}

		case elevator_floor := <-floor_changes_chan:		//TODO: Implement floor_changes_chan
			if elevator_floor == -1{
				break
			}
			last_visited_floor = elevator_floor

			if len(requests_downward[elevator_floor]) != 0 || len(requests_upward[request_to_execute.Floor]) != 0{
				elev.Elev_set_motor_direction(elev.MOTOR_DIRECTION_STOP)
				//send requests completed
				//delete requests
			}

		default:
			//figure out which direction to move
						
	}
}
