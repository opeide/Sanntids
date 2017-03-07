package request_distributor

import (
	"../message_structs"
	"../network/peers"
	"../hardware_interface"
	"fmt"
)

var all_upward_requests [hardware_interface.N_FLOORS]message_structs.Request
var all_downward_requests [hardware_interface.N_FLOORS]message_structs.Request
var all_elevator_states = make(map[string]message_structs.Elevator_state)

var zero_request message_structs.Request = message_structs.Request{}

func Distribute_requests(
	local_id string,
	peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan message_structs.Request,
	network_request_tx_chan chan<- message_structs.Request,
	local_elevator_state_changes_tx_chan chan<- message_structs.Elevator_state,
	non_local_elevator_state_changes_rx_chan <-chan message_structs.Elevator_state,
	button_request_chan <-chan message_structs.Request,
	requests_to_execute_chan chan<- message_structs.Request,
	executed_requests_chan <-chan message_structs.Request, 
	set_lamp_chan chan<- message_structs.Set_lamp_message, 
	local_elevator_state_changes_chan <-chan message_structs.Elevator_state) {

	for {
		select {
		case button_request := <-button_request_chan:
			button_request.Primary_responsible_elevator = decide_primary_responsible_elevator(local_id)
			if button_request.Primary_responsible_elevator == local_id{
				requests_to_execute_chan <- button_request			
			}
			button_request.Message_origin_id = local_id
			network_request_tx_chan <- button_request
			
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
			set_lamp_message.Floor = executed_request.Floor
			set_lamp_message.Value = 0

			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
			set_lamp_chan <- set_lamp_message
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
			set_lamp_chan <- set_lamp_message
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
			set_lamp_chan <- set_lamp_message

			executed_request.Message_origin_id = local_id
			network_request_tx_chan <- executed_request

		case non_local_request := <-network_request_rx_chan:
			if non_local_request.Message_origin_id == local_id {break}

			if non_local_request.Is_completed{
				fmt.Println("Distributor: Received non-local completed request: ", non_local_request)
			}else{
				fmt.Println("Distributor: Received non-local non-completed request: ", non_local_request)
				if non_local_request.Primary_responsible_elevator == local_id{
					fmt.Println("Distributor: Executor should do non-local request")
					non_local_request.Message_origin_id = local_id
					network_request_tx_chan <- non_local_request
				}else{
					fmt.Println("Distributor: Executor should *not* do non-local request")
				}
			}

		case local_elevator_state := <-local_elevator_state_changes_chan:
			fmt.Println("Distributor: New local elevator state: ", local_elevator_state)
			local_elevator_state.Elevator_id = local_id
			all_elevator_states[local_id] = local_elevator_state
			local_elevator_state_changes_tx_chan <- local_elevator_state

		case non_local_elevator_state := <- non_local_elevator_state_changes_rx_chan:
			if non_local_elevator_state.Elevator_id == local_id {
				break
			}

			all_elevator_states[non_local_elevator_state.Elevator_id] = non_local_elevator_state
			fmt.Println("Distributor: New non-local elevator state: ", non_local_elevator_state)

		case peer_update := <-peer_update_chan:
			if peer_update.New != "" {
				fmt.Println("Distributor: New Peer: ", peer_update.New)
			}

			if len(peer_update.Lost) != 0{
				for _, lost_elevator_id := range peer_update.Lost {
					if lost_elevator_id == local_id {
						continue
					}

					delete(all_elevator_states, lost_elevator_id)
					fmt.Println("Distributor: Deleted ", lost_elevator_id, " from elevator states. ")
				}
			}
		}
	}
}


func decide_primary_responsible_elevator(local_id string) string{
	return local_id
}