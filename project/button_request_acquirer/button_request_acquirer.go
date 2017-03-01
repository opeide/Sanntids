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
						button_request := request.Request{Floor: floor, Request_type: button_type}
						button_request_chan <- button_request
					}
					button_states[floor][button_type] = button_is_pressed
				}
			}
		}
	}
}
