package button_request_acquirer

import (
	"../elev"
	"../request"
)

var button_states [elev.N_FLOORS][3]int // 3 is the number of button types: UP, DOWN and COMMAND

func Acquire_button_requests(button_request_chan chan<- request.Request) {
	for {
		for floor := 0; floor < elev.N_FLOORS; floor++ {
			for button_type := 0; button_type < 3; button_type++ { // See elev.c for type definitions
				button_is_pressed := elev.Elev_get_button_signal(button_type, floor)
				button_was_pressed := button_states[floor][button_type]
				if button_is_pressed != button_was_pressed {
					if button_was_pressed == 1 {
						// Button released
						var direction int
						switch button_type {
						case elev.BUTTON_TYPE_CALL_UP:
							direction = elev.MOTOR_DIRECTION_UP
						case elev.BUTTON_TYPE_CALL_DOWN:
							direction = elev.MOTOR_DIRECTION_DOWN
						case elev.BUTTON_TYPE_COMMAND:
							direction = elev.MOTOR_DIRECTION_STOP
						}
						button_request := request.Request{Floor: floor, Direction: direction}
						button_request_chan <- button_request
					}
					button_states[floor][button_type] = button_is_pressed
				}
			}
		}
	}
}
