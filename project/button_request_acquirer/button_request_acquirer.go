package button_request_acquirer

import (
	"../elev"
	"../global"
)

var button_states [elev.N_FLOORS][3]int // 3 is the number of button types: UP, DOWN and COMMAND

func Acquire_button_requests(button_request_chan chan<- global.Request) {
	for {
		for floor := 1; floor <= elev.N_FLOORS; floor++ {
			for button_type := 0; button_type < 3; button_type++ { // See elev.c for type definitions
				state := elev.Elev_get_button_signal(button_type, floor)
				last_state := button_states[floor-1][button_type]
				if state != last_state {
					if last_state == 1 {
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
						button_request := global.Request{Floor: floor, Direction: direction}
						button_request_chan <- button_request
					}
					button_states[floor-1][button_type] = state
				}
			}
		}
	}
}
