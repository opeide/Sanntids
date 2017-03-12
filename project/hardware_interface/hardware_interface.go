package hardware_interface

import (
	"../message_structs"
	//"fmt"
)

const (
	LAMP_TYPE_UP = iota
	LAMP_TYPE_DOWN
	LAMP_TYPE_COMMAND
	LAMP_TYPE_FLOOR_INDICATOR
	LAMP_TYPE_DOOR_OPEN
)

var button_states [N_FLOORS][3]int // 3 is the number of button types: UP, DOWN and COMMAND
var floor_sensor_state int = -1

func Read_and_write_to_hardware(
	button_request_chan chan<- message_structs.Request,
	floor_changes_chan chan<- int,
	set_motor_direction_chan <-chan int,
	set_lamp_chan <-chan message_structs.Set_lamp_message) {

	elev_init()
	floor_sensor_reading := floor_sensor_multiple_read()
	floor_sensor_state = floor_sensor_reading
	floor_changes_chan <- floor_sensor_reading

	go button_request_acquirer(button_request_chan)
	go floor_sensor(floor_changes_chan)
	go motor_direction_setter(set_motor_direction_chan)
	go lamp_setter(set_lamp_chan)
}

func button_request_acquirer(button_request_chan chan<- message_structs.Request) {
	for {
		for floor := 0; floor < N_FLOORS; floor++ {
			for button_type := 0; button_type < 3; button_type++ { // See elev.c for type definitions
				button_is_pressed := elev_get_button_signal(button_type, floor)
				button_was_pressed := button_states[floor][button_type]
				if button_is_pressed != button_was_pressed {
					if button_was_pressed == 1 {
						// Button released
						button_request := message_structs.Request{Floor: floor, Request_type: button_type}
						button_request_chan <- button_request
					}
					button_states[floor][button_type] = button_is_pressed
				}
			}
		}
	}
}

func floor_sensor(floor_changes_chan chan<- int) {
	for {
		floor_sensor_reading := floor_sensor_multiple_read()
		if floor_sensor_reading != floor_sensor_state {
			floor_sensor_state = floor_sensor_reading
			floor_changes_chan <- floor_sensor_state
		}
	}
}
func floor_sensor_multiple_read() int {
	last_read := elev_get_floor_sensor_signal()
	for {
		floor_sensor_reading1 := elev_get_floor_sensor_signal()
		floor_sensor_reading2 := elev_get_floor_sensor_signal()
		floor_sensor_reading3 := elev_get_floor_sensor_signal()
		if floor_sensor_reading1 == last_read &&
			floor_sensor_reading2 == last_read &&
			floor_sensor_reading3 == last_read {

			return last_read
		} else {
			last_read = floor_sensor_reading3
		}
	}
}

func motor_direction_setter(set_motor_direction_chan <-chan int) {
	for {
		select {
		case new_motor_direction := <-set_motor_direction_chan:
			elev_set_motor_direction(new_motor_direction)
		}
	}
}

func lamp_setter(set_lamp_chan <-chan message_structs.Set_lamp_message) {
	for {
		select {
		case set_lamp_message := <-set_lamp_chan:
			switch set_lamp_message.Lamp_type {
			case LAMP_TYPE_UP:
				elev_set_button_lamp(BUTTON_TYPE_CALL_UP, set_lamp_message.Floor, set_lamp_message.Value)
				break
			case LAMP_TYPE_DOWN:
				elev_set_button_lamp(BUTTON_TYPE_CALL_DOWN, set_lamp_message.Floor, set_lamp_message.Value)
				break
			case LAMP_TYPE_COMMAND:
				elev_set_button_lamp(BUTTON_TYPE_COMMAND, set_lamp_message.Floor, set_lamp_message.Value)
				break
			case LAMP_TYPE_FLOOR_INDICATOR:
				elev_set_floor_indicator(set_lamp_message.Floor)
				break
			case LAMP_TYPE_DOOR_OPEN:
				elev_set_door_open_lamp(set_lamp_message.Value)
				break
			}
		}
	}
}
