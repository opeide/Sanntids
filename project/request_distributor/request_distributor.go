package request_distributor

import (
	"../message_structs"
	"../network/peers"
	"../hardware_interface"
	"fmt"
)

func Distribute_requests(id string,
	peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan message_structs.Request,
	network_request_tx_chan chan<- message_structs.Request,
	button_request_chan <-chan message_structs.Request,
	requests_to_execute_chan chan<- message_structs.Request,
	executed_requests_chan <-chan message_structs.Request, 
	set_lamp_chan chan<- message_structs.Set_lamp_message) {

	for {
		select {
		case button_request := <-button_request_chan:
			button_request.Primary_responsible_elevator = id
			requests_to_execute_chan <- button_request
			
			set_lamp_message := message_structs.Set_lamp_message{}
			switch(button_request.Request_type){
			case hardware_interface.BUTTON_TYPE_CALL_UP:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
				break
			case hardware_interface.BUTTON_TYPE_CALL_DOWN:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
				break
			case hardware_interface.BUTTON_TYPE_COMMAND:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
				break
			}
			set_lamp_message.Floor = button_request.Floor
			set_lamp_message.Value = 1
			set_lamp_chan <- set_lamp_message
		case executed_request := <-executed_requests_chan:
			fmt.Println("Distributor: Executor completed request: ", executed_request)
			
			set_lamp_message := message_structs.Set_lamp_message{}
			switch(executed_request.Request_type){
			case hardware_interface.BUTTON_TYPE_CALL_UP:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
				break
			case hardware_interface.BUTTON_TYPE_CALL_DOWN:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
				break
			case hardware_interface.BUTTON_TYPE_COMMAND:
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
				break
			}
			set_lamp_message.Floor = executed_request.Floor
			set_lamp_message.Value = 0
			set_lamp_chan <- set_lamp_message
		}
	}
}
