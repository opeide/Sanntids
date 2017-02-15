package elev

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

func Elev_init() {
	C.elev_init()
}

func Elev_set_motor_direction(motor_direction int) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(motor_direction))
}
