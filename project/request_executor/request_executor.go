package request_executor

import (
	"../hardware_interface"
	"../message_structs"
	"time"
	"fmt"
)

var last_motor_direction int
var current_elevator_floor int
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

var requests_upward [hardware_interface.N_FLOORS] message_structs.Request
var requests_downward [hardware_interface.N_FLOORS] message_structs.Request

func Execute_requests(	requests_to_execute_chan <-chan message_structs.Request, 
						executed_requests_chan chan<- message_structs.Request, 
						floor_changes_chan <-chan int,
						set_motor_direction_chan chan<- int) {
	
	elevator_initialize_position()
	
	for {
		select {
		case request_to_execute := <-requests_to_execute_chan:
			fmt.Println("executor got request.")
			switch request_to_execute.Direction {
				case hardware_interface.BUTTON_TYPE_CALL_DOWN:
					if len(requests_downward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}
				case hardware_interface.BUTTON_TYPE_CALL_UP:
					if len(requests_upward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}	
				case hardware_interface.BUTTON_TYPE_COMMAND:
					if len(requests_downward[request_to_execute.Floor]) == 0{
						requests_downward[request_to_execute.Floor] = request_to_execute
					}
					if len(requests_upward[request_to_execute.Floor]) == 0{
						requests_upward[request_to_execute.Floor] = request_to_execute
					}	
			}

			elevator_complete_request_at_current_floor()

			elevator_move_in_correct_direction()
		}

		case current_elevator_floor := <-floor_changes_chan:
			fmt.Println("executor got floor ", current_elevator_floor)	
			if current_elevator_floor == -1 {break}

			last_elevator_floor = current_elevator_floor
			elevator_complete_request_at_current_floor()
			elevator_move_in_correct_direction()						
	}
}
	
func elevator_initialize_position(){
	set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_DOWN)
	select{
		case current_elevator_floor := <-floor_changes_chan:
			if current_elevator_floor == -1 {break}
			last_visited_floor = current_elevator_floor
			return
		case <-time.After(time.Second * 5):
			set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_UP)
	}
	select{
		case current_elevator_floor := <-floor_changes_chan:
			if current_elevator_floor == -1 {break}
			last_visited_floor = current_elevator_floor
			return
		case <-time.After(time.Second * 5):
			fmt.Println("ELEVATOR DID NOT FIND ANY FLOORS DURING INIT.")
	}

}
func elevator_complete_request_at_current_floor(){
	if current_elevator_floor == -1 {break}
	if len(requests_downward[elevator_floor]) != 0{
		set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_STOP)	//TODO: Turn off lights and open doors
		executed_requests_chan <- requests_downward[elevator_floor]
		requests_downward[elevator_floor] = nil 
	}

	if len(requests_upward[elevator_floor]) != 0{
		set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_STOP)	//TODO: Turn off lights and open doors
		executed_requests_chan <- requests_upward[elevator_floor]
		requests_upward[elevator_floor] = nil
	}
}	

func elevator_move_in_correct_direction(){
	if elevator_floor == -1 {break}

	if last_motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN{
		has_requests_below := 0
		for (i = 0; i < last_elevator_floor; i++){
			has_requests_below += len(requests_downward[i])
		}
		if has_requests_below{
			set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_DOWN)
		}
	}
	
	has_requests_above := 0
	for (i = last_elevator_floor+1; i < hardware_interface.N_FLOORS; i++){
		has_requests_above, += len(requests_downward[i])
	}
	if has_requests_above{
		set_motor_direction_chan <- (hardware_interface.MOTOR_DIRECTION_UP)
	}
}