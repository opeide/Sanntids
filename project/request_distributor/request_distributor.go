package request_distributor

import (
	"../hardware_interface"
	"../message_structs"
	"../network/peers"
	"fmt"
)

var all_upward_requests map[string][hardware_interface.N_FLOORS]message_structs.Request
var all_downward_requests map[string][hardware_interface.N_FLOORS]message_structs.Request
var all_command_requests map[string][hardware_interface.N_FLOORS]message_structs.Request

var zero_request message_structs.Request = message_structs.Request{}

var all_elevator_states = make(map[string]message_structs.Elevator_state)

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
			button_request.Primary_responsible_elevator = decide_responsible_elevator(local_id, button_request)
			if button_request.Primary_responsible_elevator == local_id {
				requests_to_execute_chan <- button_request
				set_request_lights(local_id, set_lamp_chan, button_request, 1)
			}
			store_request(button_request)
			button_request.Message_origin_id = local_id
			network_request_tx_chan <- button_request

		case executed_request := <-executed_requests_chan:
			fmt.Println("Distributor: Executor completed request: ", executed_request)

			set_request_lights(local_id, set_lamp_chan, executed_request, 0)

			executed_request.Message_origin_id = local_id
			network_request_tx_chan <- executed_request

		case non_local_request := <-network_request_rx_chan:
			if non_local_request.Message_origin_id == local_id {
				break
			}

			if non_local_request.Is_completed {
				fmt.Println("Distributor: Received non-local completed request: ", non_local_request)
				set_request_lights(local_id, set_lamp_chan, non_local_request, 0)
				delete_request_and_related(non_local_request)
			} else {
				store_request(non_local_request)
				fmt.Println("Distributor: Received non-local non-completed request: ", non_local_request)
				set_request_lights(local_id, set_lamp_chan, non_local_request, 1)

				if non_local_request.Primary_responsible_elevator == local_id {
					fmt.Println("Distributor: Executor should do non-local request")
					requests_to_execute_chan <- non_local_request
					non_local_request.Message_origin_id = local_id
					network_request_tx_chan <- non_local_request
				} else {
					fmt.Println("Distributor: Executor should *not* do non-local request")
				}
			}

		case local_elevator_state := <-local_elevator_state_changes_chan:
			fmt.Println("Distributor: New local elevator state: ", local_elevator_state)
			local_elevator_state.Elevator_id = local_id
			all_elevator_states[local_id] = local_elevator_state
			local_elevator_state_changes_tx_chan <- local_elevator_state

		case non_local_elevator_state := <-non_local_elevator_state_changes_rx_chan:
			if non_local_elevator_state.Elevator_id == local_id {
				break
			}

			all_elevator_states[non_local_elevator_state.Elevator_id] = non_local_elevator_state
			fmt.Println("Distributor: New non-local elevator state: ", non_local_elevator_state)

		case peer_update := <-peer_update_chan:
			if peer_update.New != "" {
				fmt.Println("Distributor: New Peer: ", peer_update.New)
				if peer_update.New != local_id {
					local_elevator_state_changes_tx_chan <- all_elevator_states[local_id]
				} else {
					send_all_requests_to_network(network_request_tx_chan)
				}
			}

			if len(peer_update.Lost) != 0 {
				for _, lost_elevator_id := range peer_update.Lost {
					if lost_elevator_id == local_id {
						continue
					} else {
						execute_requests_belonging_to(lost_elevator_id, requests_to_execute_chan)
					}

					delete(all_elevator_states, lost_elevator_id)
					fmt.Println("Distributor: Deleted ", lost_elevator_id, " from elevator states. ")
				}
			}
		}
	}
}

//all_upward_requests map[string][hardware_interface.N_FLOORS]message_structs.Request
func send_all_requests_to_network(network_request_tx_chan chan<- message_structs.Request) {
	// This for each loop makes copies of the arrays, not references.
	for _, requests_by_id := range []map[string][hardware_interface.N_FLOORS]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
		for _, request_by_floor := range requests_by_id {
			for _, request := range request_by_floor {
				if request != zero_request {
					network_request_tx_chan <- request
				}
			}
		}
	}
}

func execute_requests_belonging_to(elevator_id string, requests_to_execute_chan chan<- message_structs.Request) {
	for _, requests_by_id := range []map[string][hardware_interface.N_FLOORS]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
		for loop_id, request_by_floor := range requests_by_id {
			if loop_id != elevator_id {
				continue
			}
			for _, request := range request_by_floor {
				if request != zero_request {
					requests_to_execute_chan <- request
				}
			}
		}
	}
}

func store_request(request message_structs.Request) {
	/*
		switch request.Request_type{
			case BUTTON_TYPE_CALL_UP:

			case BUTTON_TYPE_CALL_DOWN:

			case BUTTON_TYPE_COMMAND:

		}
	*/
}

func delete_request_and_related(request message_structs.Request) {
	/*
		switch request.Request_type{
			case BUTTON_TYPE_CALL_UP:

			case BUTTON_TYPE_CALL_DOWN:

			case BUTTON_TYPE_COMMAND:

		}
	*/
}

func set_request_lights(local_id string,
	set_lamp_chan chan<- message_structs.Set_lamp_message,
	request message_structs.Request,
	turn_on int) {

	set_lamp_message := message_structs.Set_lamp_message{}
	set_lamp_message.Floor = request.Floor
	set_lamp_message.Value = turn_on
	if turn_on == 1 {
		switch request.Request_type {
		case hardware_interface.BUTTON_TYPE_CALL_UP:
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
			break
		case hardware_interface.BUTTON_TYPE_CALL_DOWN:
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
			break
		case hardware_interface.BUTTON_TYPE_COMMAND:
			if request.Primary_responsible_elevator == local_id {
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
			}
			break
		}
		set_lamp_chan <- set_lamp_message
	} else {
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
		set_lamp_chan <- set_lamp_message
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
		set_lamp_chan <- set_lamp_message
		if request.Primary_responsible_elevator == local_id {
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
			set_lamp_chan <- set_lamp_message
		}
	}
}

func decide_responsible_elevator(local_id string, request message_structs.Request) string {
	if request.Request_type == hardware_interface.BUTTON_TYPE_COMMAND {
		return local_id
	}

	fastest_time := estimate_time_for_elevator_to_complete_request(all_elevator_states[local_id], request)
	fastest_elevator_id := local_id

	for elevator_id, elevator_state := range all_elevator_states {
		time := estimate_time_for_elevator_to_complete_request(elevator_state, request)
		if time < fastest_time {
			fastest_time = time
			fastest_elevator_id = elevator_id
		}
	}
	fmt.Println(fastest_elevator_id, " would solve request fastest: ", request)
	return fastest_elevator_id
}

func estimate_time_for_elevator_to_complete_request(elevator_state message_structs.Elevator_state, request message_structs.Request) int {
	time := abs(elevator_state.Last_visited_floor-request.Floor) * 5
	return time
}

func abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}
