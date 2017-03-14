package hardware_interface

/*
#cgo CFLAGS: -std=c99
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

const (
	N_FLOORS  = C.N_FLOORS
	N_BUTTONS = C.N_BUTTONS

	MOTOR_DIRECTION_DOWN = int(C.int(C.elev_motor_direction_t(C.DIRN_DOWN)))
	MOTOR_DIRECTION_STOP = int(C.int(C.elev_motor_direction_t(C.DIRN_STOP)))
	MOTOR_DIRECTION_UP   = int(C.int(C.elev_motor_direction_t(C.DIRN_UP)))

	BUTTON_TYPE_CALL_UP   = int(C.int(C.elev_button_type_t(C.BUTTON_CALL_UP)))
	BUTTON_TYPE_CALL_DOWN = int(C.int(C.elev_button_type_t(C.BUTTON_CALL_DOWN)))
	BUTTON_TYPE_COMMAND   = int(C.int(C.elev_button_type_t(C.BUTTON_COMMAND)))
)

func elev_init() { C.elev_init() }

func elev_set_motor_direction(motor_direction int) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(motor_direction))
}
func elev_set_button_lamp(button_type int, floor int, value int) {
	C.elev_set_button_lamp(C.elev_button_type_t(button_type), C.int(floor), C.int(value))
}
func elev_set_floor_indicator(floor int) { C.elev_set_floor_indicator(C.int(floor)) }
func elev_set_door_open_lamp(value int)  { C.elev_set_door_open_lamp(C.int(value)) }
func elev_set_stop_lamp(value int)       { C.elev_set_stop_lamp(C.int(value)) }

func elev_get_button_signal(button_type int, floor int) int {
	return int(C.int(C.elev_get_button_signal(C.elev_button_type_t(button_type), C.int(floor))))
}
func elev_get_floor_sensor_signal() int { return int(C.int(C.elev_get_floor_sensor_signal())) }
func elev_get_stop_signal() int         { return int(C.int(C.elev_get_stop_signal())) }
func elev_get_obstruction_signal() int  { return int(C.int(C.elev_get_obstruction_signal())) }
