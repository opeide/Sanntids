package hardware_interface

/*
#cgo CFLAGS: -std=c99
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"

// Returns 0 on init failure
func io_init() int {
	return int(C.int(C.io_init()))
}

func io_set_bit(channel int) {
	C.io_set_bit(C.int(channel))
}

func io_clear_bit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func io_read_bit(channel int) int {
	return int(C.int(C.io_read_bit(C.int(channel))))
}

func io_read_analog(channel int) int {
	return int(C.int(C.io_read_analog(C.int(channel))))
}

func io_write_analog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}
