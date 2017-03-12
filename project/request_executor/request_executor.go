package request_executor

import (
	"../hardware_interface"
	"../message_structs"
	"../motor_movement_watchdog"
	"fmt"
	"os"
	"time"
)

const (
	STATE_TYPE_IDLE = iota
	STATE_TYPE_DOORS_OPEN
	STATE_TYPE_MOVING_DOWN
	STATE_TYPE_MOVING_UP
)

var current_elevator_state_type int
var last_non_stop_motor_direction int
var last_visited_floor int

var door_just_closed_chan chan int = make(chan int, 1)

const DOOR_OPEN_TIME = 2                // Seconds
const INITIALIZATION_TIMEOUT = 5        // Seconds
const ESTIMATED_TIME_BETWEEN_FLOORS = 2 //seconds

var requests_upward [hardware_interface.N_FLOORS]message_structs.Request
var requests_downward [hardware_interface.N_FLOORS]message_structs.Request
var requests_command [hardware_interface.N_FLOORS]message_structs.Request
var zero_request message_structs.Request = message_structs.Request{}

var executed_requests_chan chan<- message_structs.Request
var set_lamp_chan chan<- message_structs.Set_lamp_message
var set_motor_direction_chan chan<- int
var local_elevator_state_changes_chan chan<- message_structs.Elevator_state
var floor_changes_chan <-chan int

//TODO: make a timer that measures the time in shaft. if more than a specified amount of time passes, something is [wrong]!

func Execute_requests(
	executed_requests_chan_parameter chan<- message_structs.Request,
	set_motor_direction_chan_parameter chan<- int,
	set_lamp_chan_parameter chan<- message_structs.Set_lamp_message,
	local_elevator_state_changes_chan_parameter chan<- message_structs.Elevator_state,
	floor_changes_chan_parameter <-chan int,
	requests_to_execute_chan <-chan message_structs.Request) {

	executed_requests_chan = executed_requests_chan_parameter
	floor_changes_chan = floor_changes_chan_parameter
	set_motor_direction_chan = set_motor_direction_chan_parameter
	set_lamp_chan = set_lamp_chan_parameter
	local_elevator_state_changes_chan = local_elevator_state_changes_chan_parameter

	elevator_initialize_position()

	for {
		select {
		case current_floor := <-floor_changes_chan:
			if current_floor == -1 {
				if current_elevator_state_type == STATE_TYPE_IDLE {
					fmt.Println("ILLEGAL STATE: IDLE IN SHAFT. Exiting...")
					os.Exit(0) //Lets backup take over (effectively a program restart)
				}
				break
			}

			motor_movement_watchdog.Timer_stop()

			set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_FLOOR_INDICATOR, Floor: current_floor}
			last_visited_floor = current_floor
			set_next_correct_state_being_at(last_visited_floor, last_non_stop_motor_direction)

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
				if requests_command[request_to_execute.Floor] == zero_request {
					requests_command[request_to_execute.Floor] = request_to_execute
				}
			}

			if current_elevator_state_type == STATE_TYPE_IDLE {
				if request_to_execute.Floor == last_visited_floor {
					set_state(STATE_TYPE_DOORS_OPEN, last_visited_floor, last_non_stop_motor_direction)
				} else if request_to_execute.Floor < last_visited_floor {
					set_state(STATE_TYPE_MOVING_DOWN, last_visited_floor, hardware_interface.MOTOR_DIRECTION_DOWN)
				} else { // request_to_execute.Floor > last_visited_floor
					set_state(STATE_TYPE_MOVING_UP, last_visited_floor, hardware_interface.MOTOR_DIRECTION_UP)
				}
			}

		case <-door_just_closed_chan:
			register_finished_requests_at(last_visited_floor)
			set_next_correct_state_being_at(last_visited_floor, last_non_stop_motor_direction)
		}
	}
}

func elevator_initialize_position() {
	select {
	case current_floor := <-floor_changes_chan:
		if current_floor != -1 {
			set_state(STATE_TYPE_IDLE, current_floor, hardware_interface.MOTOR_DIRECTION_DOWN)
			return
		}
	}

	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_DOWN
	select {
	case current_floor := <-floor_changes_chan:
		set_state(STATE_TYPE_IDLE, current_floor, hardware_interface.MOTOR_DIRECTION_DOWN)
		return
	case <-time.After(time.Second * INITIALIZATION_TIMEOUT):
		break
	}

	fmt.Println("ELEVATOR DID NOT FIND ANY FLOORS DURING EXECUTOR INIT. Exiting...")
	set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
	os.Exit(0) //Lets backup take over (effectively a program restart)
}

// Should only be called when at a floor and finished in state type STATE_TYPE_DOORS_OPEN
func set_next_correct_state_being_at(current_floor int, current_direction int) {
	// Should we have this test?
	/*if current_floor == -1 {
		switch current_direction {
		case hardware_interface.MOTOR_DIRECTION_DOWN:
			return STATE_TYPE_MOVING_DOWN
		case hardware_interface.MOTOR_DIRECTION_UP:
			return STATE_TYPE_MOVING_UP
		}
	}*/

	if is_request_at(current_floor, current_direction) {
		set_state(STATE_TYPE_DOORS_OPEN, current_floor, current_direction)
		return
	}

	if has_request_in_direction(current_floor, current_direction) {
		new_state := motor_direction_to_directional_state(current_direction)
		set_state(new_state, current_floor, current_direction)
		return
	}

	if is_request_at(current_floor, -current_direction) {
		set_state(STATE_TYPE_DOORS_OPEN, current_floor, -current_direction) // Is why we need current_direction argument
		return
	}

	if has_request_in_direction(current_floor, -current_direction) {
		new_state := motor_direction_to_directional_state(-current_direction)
		set_state(new_state, current_floor, -current_direction)
		return
	}

	set_state(STATE_TYPE_IDLE, current_floor, current_direction)
}

// Should not be called while not finished in state type DOORS_OPEN
func set_state(new_state_type int, new_last_visited_floor int, new_last_non_stop_direction int) {
	set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_DOOR_OPEN, Value: 0}

	switch new_state_type {
	case STATE_TYPE_IDLE:
		select {
		case <-time.After(time.Millisecond * 100): // Place close to the middel of the sensor
		}
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
		set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_FLOOR_INDICATOR, Floor: new_last_visited_floor}

	case STATE_TYPE_DOORS_OPEN:
		set_motor_direction_chan <- hardware_interface.MOTOR_DIRECTION_STOP
		set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_FLOOR_INDICATOR, Floor: new_last_visited_floor}

		set_lamp_chan <- message_structs.Set_lamp_message{Lamp_type: hardware_interface.LAMP_TYPE_DOOR_OPEN, Value: 1}
		go func() {
			select {
			case <-time.After(time.Second * DOOR_OPEN_TIME):
				door_just_closed_chan <- 1
			}
		}()

	case STATE_TYPE_MOVING_DOWN:
		new_last_non_stop_direction = hardware_interface.MOTOR_DIRECTION_DOWN
		set_motor_direction_chan <- new_last_non_stop_direction
		motor_movement_watchdog.Timer_start()

	case STATE_TYPE_MOVING_UP:
		new_last_non_stop_direction = hardware_interface.MOTOR_DIRECTION_UP
		set_motor_direction_chan <- new_last_non_stop_direction
		motor_movement_watchdog.Timer_start()
	}

	last_non_stop_motor_direction = new_last_non_stop_direction
	last_visited_floor = new_last_visited_floor
	current_elevator_state_type = new_state_type

	local_elevator_state_changes_chan <- message_structs.Elevator_state{
		Last_visited_floor:            last_visited_floor,
		Last_non_stop_motor_direction: last_non_stop_motor_direction}
}

// Includes if there is a COMMAND request there
func is_request_at(floor int, motor_direction int) bool {
	if requests_command[floor] != zero_request {
		return true
	}

	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		if requests_downward[floor] != zero_request {
			return true
		}
	}

	if motor_direction == hardware_interface.MOTOR_DIRECTION_UP {
		if requests_upward[floor] != zero_request {
			return true
		}
	}

	return false
}

func register_finished_requests_at(floor int) {
	if requests_downward[floor] != zero_request {
		requests_downward[floor].Is_completed = true
		executed_requests_chan <- requests_downward[floor]
		requests_downward[floor] = zero_request
	}
	if requests_upward[floor] != zero_request {
		requests_upward[floor].Is_completed = true
		executed_requests_chan <- requests_upward[floor]
		requests_upward[floor] = zero_request
	}
	if requests_command[floor] != zero_request {
		requests_command[floor].Is_completed = true
		executed_requests_chan <- requests_command[floor]
		requests_command[floor] = zero_request
	}
}

func has_request_in_direction(floor int, motor_direction int) bool {
	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		if floor == 0 {
			return false
		}

		for floor_below := 0; floor_below < floor; floor_below++ {
			if requests_downward[floor_below] != zero_request ||
				requests_upward[floor_below] != zero_request ||
				requests_command[floor_below] != zero_request {
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
			if requests_downward[floor_above] != zero_request ||
				requests_upward[floor_above] != zero_request ||
				requests_command[floor_above] != zero_request {
				return true
			}
		}
		return false
	}
	return false
}

func motor_direction_to_directional_state(motor_direction int) int {
	if motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {
		return STATE_TYPE_MOVING_DOWN
	} else if motor_direction == hardware_interface.MOTOR_DIRECTION_UP {
		return STATE_TYPE_MOVING_UP
	} else { // motor_direction == hardware_interface.MOTOR_DIRECTION_STOP {
		return STATE_TYPE_IDLE
	}
}
