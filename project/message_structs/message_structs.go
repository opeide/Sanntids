package message_structs

type Request struct {
	Floor                          int
	Request_type                   int // Same as button type
	Primary_responsible_elevator   string
	Secondary_responsible_elevator string
	Is_completed                   bool
}

type Set_lamp_message struct {
	Lamp_type int
	Floor     int
	Value     int
}
