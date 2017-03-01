package elev_io

/*
#cgo CFLAGS: -std=c99
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

// Returns 0 on init failure
func IO_init() int {
	return int(C.int(C.io_init()))
}

func IO_set_bit(channel int) {
	C.io_set_bit(C.int(channel))
}

func IO_clear_bit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func IO_read_bit(channel int) int {
	return int(C.int(io_read_bit(C.int(channel))))
}

func IO_read_analog(channel int) int {
	return int(C.int(C.io_read_analog(C.int(channel))))
}

func IO_write_analog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}