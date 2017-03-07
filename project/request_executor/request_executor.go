package request_executor

import (
	"../hardware_interface"
	"../message_structs"
	"fmt"
	"time"
)

var last_non_stop_motor_direction int = hardware_interface.MOTOR_DIRECTION_UP
var current_floor int //-1 if not at floor
var last_visited_floor int
var door_is_open bool = false

var door_just_closed_chan chan int = make(chan int, 1)

var requests_upward [hardware_interface.N_FLOORS]message_structs.Request
var requests_downward [hardware_interface.N_FLOORS]message_structs.Request
var zero_request message_structs.Request = message_structs.Request{} //dette ok?????????????????????????????????????????????????????????????????????????????????????????????

func Execute_requests(
	requests_to_execute_chan <-chan message_structs.Request,
	executed_requests_chan chan<- message_structs.Request,
	floor_changes_chan <-chan int,
	set_motor_direction_chan chan<- int, 
	set_lamp_chan chan<- message_structs.Set_lamp_message, 
	local_elevator_state_changes_chan chan<- message_structs.Elevator_state) {

	elevator_initialize_position(set_motor_direction_chan, floor_changes_chan)
	set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_FLOOR_INDICATOR, Floor: current_floor}

	for {
		select {
		case request_to_execute := <-requests_to_execute_chan:
			switch request_to_execute.Request_type {
			case hardware_interface.BUTTON_TYPE_CALL_DOWN:
				if requests_downward[request_to_execute.Floor] == zero_request {
					requests_downward[request_to_execute.Floor] = request_to_execute
				}
			case hardware_interface.BUTTON_TYPE_CALL_UP:
				if requests_upward[request_to_execute.Floor] == zero_request {
					requests_upward[request_to_execute.Floor] = request_to_execute
				}
			case hardware_interface.BUTTON_TYPE_COMMAND:
				if requests_downward[request_to_execute.Floor] == zero_request {
					requests_downward[request_to_execute.Floor] = request_to_execute
				}
				if requests_upward[request_to_execute.Floor] == zero_request {
					requests_upward[request_to_execute.Floor] = request_to_execute
				}
			}
			
			elevator_attempt_complete_request_at_current_floor(set_motor_direction_chan, set_lamp_chan, executed_requests_chan)
			elevator_move_in_correct_direction(set_motor_direction_chan)

		case current_floor = <-floor_changes_chan:
			if current_floor == -1 {
				break
			}

			set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_FLOOR_INDICATOR, Floor: current_floor}			
			last_visited_floor = current_floor	
			elevator_attempt_complete_request_at_current_floor(set_motor_direction_chan, set_lamp_chan, executed_requests_chan)
			elevator_move_in_correct_direction(set_motor_direction_chan)

			local_elevator_state_changes_chan <- message_structs.Elevator_state{ 
				Last_visited_floor: last_visited_floor,
				Last_non_stop_motor_direction: last_non_stop_motor_direction}

		case <-door_just_closed_chan:
			door_is_open = false
			elevator_attempt_complete_request_at_current_floor(set_motor_direction_chan, set_lamp_chan, executed_requests_chan)
			elevator_move_in_correct_direction(set_motor_direction_chan)
		}
	}
}

func elevator_initialize_position(
	set_motor_direction_chan chan<- int,
	floor_changes_chan <-chan int) {

	select {
	case current_floor = <-floor_changes_chan:
		if current_floor != -1 {
			last_visited_floor = current_floor
			set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP //precautionary measure, do not trust low level init files
			return
		}
	}

	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_DOWN
	select {
	case current_floor = <-floor_changes_chan:
		last_visited_floor = current_floor
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
		return
	case <-time.After(time.Second * 5):
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_UP
	}
	select {
	case current_floor = <-floor_changes_chan:
		last_visited_floor = current_floor
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
		return
	case <-time.After(time.Second * 5):
		break
	}
	fmt.Println("ELEVATOR DID NOT FIND ANY FLOORS DURING EXECUTOR INIT. SHOULD RESTART.")
	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
}

func elevator_attempt_complete_request_at_current_floor(
	set_motor_direction_chan chan<- int,
	set_lamp_chan chan<- message_structs.Set_lamp_message,
	executed_requests_chan chan<- message_structs.Request) {

	if current_floor == -1 || door_is_open{
		return
	}

	request_here := get_request_at(current_floor, last_non_stop_motor_direction)
	if request_here != zero_request {
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP //TODO: Turn off lights and open doors
		door_is_open = true
		go open_doors_and_close_after_time(set_lamp_chan)
		request_here.Is_completed = true
		executed_requests_chan <- request_here
		set_request_at(current_floor, last_non_stop_motor_direction, zero_request)
		set_request_at(current_floor, -last_non_stop_motor_direction, zero_request) // See elev.c: -1*dir is opposite dir
		return
	}

	if has_request_in_direction(current_floor, last_non_stop_motor_direction) {
		return
	}

	request_here = get_request_at(current_floor, -last_non_stop_motor_direction)
	if request_here != zero_request {
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP //TODO: Turn off lights and open doors
		last_non_stop_motor_direction = -last_non_stop_motor_direction
		door_is_open = true
		go open_doors_and_close_after_time(set_lamp_chan)
		request_here.Is_completed = true
		executed_requests_chan <- request_here
		set_request_at(current_floor, last_non_stop_motor_direction, zero_request)
		return
	}

	if has_request_in_direction(current_floor, -last_non_stop_motor_direction) {
		set_motor_direction_chan <- -last_non_stop_motor_direction
		last_non_stop_motor_direction = -last_non_stop_motor_direction
		return
	}

	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
}


func elevator_move_in_correct_direction(
	set_motor_direction_chan chan<- int) {

	if door_is_open{
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
		return
	}

	switch current_floor {
	case -1:
		set_motor_direction_chan <- last_non_stop_motor_direction
		return
	case 0, hardware_interface.N_FLOORS - 1:
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
	}

	if has_request_in_direction(current_floor, last_non_stop_motor_direction) {
		set_motor_direction_chan <- last_non_stop_motor_direction
		return
	}

	if has_request_in_direction(current_floor, -last_non_stop_motor_direction) {
		set_motor_direction_chan <- -last_non_stop_motor_direction
		last_non_stop_motor_direction = -last_non_stop_motor_direction
		return
	}

	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
}


func open_doors_and_close_after_time(set_lamp_chan chan<- message_structs.Set_lamp_message){
	set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_DOOR_OPEN, Floor: current_floor, Value: 1}
	select{
	case  <-time.After(time.Second * 3):
		set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_DOOR_OPEN, Floor: current_floor, Value: 0}
		door_just_closed_chan <- 1
	}
}

func get_request_at(floor int, motor_direction int) message_structs.Request {
	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		return requests_downward[floor]
	}

	if motor_direction == hardware_interface.MOTOR_DIRECTION_UP {
		return requests_upward[floor]
	}

	return zero_request
}


func set_request_at(floor int, motor_direction int, request message_structs.Request) {
	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		requests_downward[floor] = request
	}

	if motor_direction == hardware_interface.MOTOR_DIRECTION_UP {
		requests_upward[floor] = request
	}
}


func has_request_in_direction(floor int, motor_direction int) bool {
	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		if floor == 0 {
			return false
		}

		for floor_below := 0; floor_below < floor; floor_below++ {
			if requests_downward[floor_below] != zero_request || requests_upward[floor_below] != zero_request {
				return true
			}
		}
		return false
	}

	if motor_direction == hardware_interface.MOTOR_DIRECTION_UP {
		if floor == hardware_interface.N_FLOORS-1 {
			return false
		}

		for floor_above := floor + 1; floor_above < hardware_interface.N_FLOORS; floor_above++ {
			if requests_downward[floor_above] != zero_request || requests_upward[floor_above] != zero_request {
				return true
			}
		}
		return false
	}
	return false
}
