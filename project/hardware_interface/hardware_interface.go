package hardware_interface

import(
	"./elev"
	"../request"
)

var button_states [elev.N_FLOORS][3]int // 3 is the number of button types: UP, DOWN and COMMAND
var floor_sensor_state int = -1

func Read_and_write_to_hardware(button_request_chan chan<- request.Request
								floor_changes_chan chan<- int){
	
	elev.Elev_init(elev.ET_Comedi)

	for {
		// Read buttons
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

		// Read floor sensor
		floor_sensor_reading := elev.Elev_get_floor_sensor_signal()
		if floor_sensor_reading != floor_sensor_state {
			floor_sensor_state = floor_sensor_reading
			floor_changes_chan <- floor_sensor_state
		}
	}
}