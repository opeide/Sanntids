package message_structs

type Request struct {
	Floor                          int
	Request_type                   int // Same as button type
	Primary_responsible_elevator   string
	Secondary_responsible_elevator string
	Is_completed                   bool
}

type Set_lamp_message struct {
	Lamp_type int // See hardware_interface for values
	Floor     int
	Value     int
}

type Elevator_state struct{
	Last_visited_floor int
	Last_non_stop_motor_direction int // See hardware_interface for motor directions
}