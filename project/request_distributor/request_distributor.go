package request_distributor

import (
	"../hardware_interface"
	"../message_structs"
	"../network/peers"
	"fmt"
	"sort"
)

var all_elevator_states = make(map[string]message_structs.Elevator_state)
var all_upward_requests = make(map[string][]message_structs.Request)
var all_downward_requests = make(map[string][]message_structs.Request)
var all_command_requests = make(map[string][]message_structs.Request)

var zero_request message_structs.Request = message_structs.Request{}

var local_id string
var network_request_tx_chan 				chan<- message_structs.Request
var local_elevator_state_changes_tx_chan 	chan<- message_structs.Elevator_state
var requests_to_execute_chan 				chan<- message_structs.Request
var set_lamp_chan 							chan<- message_structs.Set_lamp_message

// Remember to do caborders when we get back after losing power!!!!!!!!!!!!!!!
// If we lose network, will we be filling up the network channel and eventually block???????

func Distribute_requests(
	local_id_parameter string,
	network_request_tx_chan_parameter 				chan<- message_structs.Request,
	local_elevator_state_changes_tx_chan_parameter 	chan<- message_structs.Elevator_state,
	requests_to_execute_chan_parameter 				chan<- message_structs.Request,
	set_lamp_chan_parameter 						chan<- message_structs.Set_lamp_message,
	peer_update_chan 								<-chan peers.PeerUpdate,
	network_request_rx_chan 						<-chan message_structs.Request,
	non_local_elevator_state_changes_rx_chan 		<-chan message_structs.Elevator_state,
	button_request_chan 							<-chan message_structs.Request,
	executed_requests_chan 							<-chan message_structs.Request,
	local_elevator_state_changes_chan 				<-chan message_structs.Elevator_state) {
	
	local_id 								= local_id_parameter
	network_request_tx_chan 				= network_request_tx_chan_parameter
	local_elevator_state_changes_tx_chan 	= local_elevator_state_changes_tx_chan_parameter
	requests_to_execute_chan 				= requests_to_execute_chan_parameter
	set_lamp_chan 							= set_lamp_chan_parameter

	for {
		select {
		case button_request := <-button_request_chan:
			button_request.Responsible_elevator = decide_responsible_elevator(local_id, button_request)
			
			register_request(button_request)
			distribute_request(button_request)

		case executed_request := <-executed_requests_chan:
			distribute_request(executed_request)
			register_finished_request(executed_request)
			
		case non_local_request := <-network_request_rx_chan:
			if non_local_request.Message_origin_id == local_id{break}
			if non_local_request.Is_completed {
				register_finished_request(non_local_request)
			}else{
				register_request(non_local_request)
				if non_local_request.Responsible_elevator == local_id {
					distribute_request(non_local_request)
				}
			}

		case local_elevator_state := <-local_elevator_state_changes_chan:
			local_elevator_state.Elevator_id = local_id
			all_elevator_states[local_id] = local_elevator_state
			local_elevator_state_changes_tx_chan <- local_elevator_state

		case non_local_elevator_state := <-non_local_elevator_state_changes_rx_chan:
			if non_local_elevator_state.Elevator_id == local_id {break}
			all_elevator_states[non_local_elevator_state.Elevator_id] = non_local_elevator_state

		case peer_update := <-peer_update_chan:
			if peer_update.New != "" {
				fmt.Println("Distributor: New Peer: ", peer_update.New)
				all_upward_requests[peer_update.New] 	= make([]message_structs.Request, hardware_interface.N_FLOORS) 
				all_downward_requests[peer_update.New] 	= make([]message_structs.Request, hardware_interface.N_FLOORS) 
				all_command_requests[peer_update.New] 	= make([]message_structs.Request, hardware_interface.N_FLOORS) 
				if peer_update.New != local_id {
					local_elevator_state_changes_tx_chan <- all_elevator_states[local_id]
				}else{
					// Just got network, send all requests on network
					for _, requests_list_by_id := range []map[string][]message_structs.Request {all_upward_requests, all_downward_requests, all_command_requests}{
						for _, request_list_by_floor := range requests_list_by_id{
							for _, request := range request_list_by_floor{
								if request != zero_request{
									distribute_request(request)
								}
							}
						}
					}
				}				
			}

			if len(peer_update.Lost) != 0 {
				for _, lost_elevator_id := range peer_update.Lost {
					if lost_elevator_id != local_id {
						//Inherit requests and delete elevator
						for _, requests_list_by_id := range []map[string][]message_structs.Request {all_upward_requests, all_downward_requests, all_command_requests}{
							for responsible_id, request_list_by_floor := range requests_list_by_id{
								if responsible_id == lost_elevator_id {
									for _, request := range request_list_by_floor{
										if request != zero_request{
											request.Responsible_elevator = local_id
											register_request(request)
											distribute_request(request)
										}
									}
									break
								}
							}
						}
						delete(all_upward_requests, lost_elevator_id)
						delete(all_downward_requests, lost_elevator_id)
						delete(all_command_requests, lost_elevator_id)
						delete(all_elevator_states, lost_elevator_id)
					}
				}
			}
		}
	}
}

func register_request(request message_structs.Request){
	switch request.Request_type{
		case hardware_interface.BUTTON_TYPE_CALL_UP:  
			if all_upward_requests[request.Responsible_elevator][request.Floor] == zero_request{
				all_upward_requests[request.Responsible_elevator][request.Floor] = request	
			}	
		case hardware_interface.BUTTON_TYPE_CALL_DOWN: 
			if all_downward_requests[request.Responsible_elevator][request.Floor] == zero_request{
				all_downward_requests[request.Responsible_elevator][request.Floor] = request	
			}	
		case hardware_interface.BUTTON_TYPE_COMMAND:
			if all_command_requests[request.Responsible_elevator][request.Floor] == zero_request{
				all_command_requests[request.Responsible_elevator][request.Floor] = request	
			}	
	}
	set_request_lights(non_local_request, 1)
	print_request_list()
}

func distribute_request(request message_structs.Request){
	if request.Responsible_elevator == local_id && request.Is_completed == false{
		requests_to_execute_chan <- request
	}
	request.Message_origin_id = local_id
	network_request_tx_chan <- request
}

func register_finished_request(request message_structs.Request){
	set_request_lights(executed_request, 0)
	all_upward_requests[request.Responsible_elevator][request.Floor] = zero_request
	all_downward_requests[request.Responsible_elevator][request.Floor] = zero_request
	if request.Primary_reponsible == local_id{
		all_command_requests[request.Responsible_elevator][request.Floor] = zero_request
	}
	print_request_list()
}

func set_request_lights(request message_structs.Request, value int){
	set_lamp_message := message_structs.Set_lamp_message{}
	set_lamp_message.Floor = request.Floor
	set_lamp_message.Value = value
	if value == 1 {
		switch request.Request_type {
		case hardware_interface.BUTTON_TYPE_CALL_UP:
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
			set_lamp_chan <- set_lamp_message
			break
		case hardware_interface.BUTTON_TYPE_CALL_DOWN:
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
			set_lamp_chan <- set_lamp_message
			break
		case hardware_interface.BUTTON_TYPE_COMMAND:
			if request.Responsible_elevator == local_id{
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
				set_lamp_chan <- set_lamp_message
			}
		}
	}else{
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
		set_lamp_chan <- set_lamp_message
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
		set_lamp_chan <- set_lamp_message

		if request.Responsible_elevator == local_id{
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

func print_request_list(){
	var ids []string
	for id := range all_command_requests{
		//fmt.Println(id)
		ids = append(ids, id)
	}
	sort.Strings(ids)
	//fmPrintln(ids)

	fmt.Print("\n\n\n\n")
	for responsible_id := range all_command_requests {
		fmt.Print("\n")
		fmt.Println("Responsible: ", responsible_id)
		fmt.Println("--------------------------------------")
		fmt.Println("\tFLOOR\tUP\tDOWN\tCOMMAND")
		for floor:= hardware_interface.N_FLOORS-1; floor >= 0 ; floor--{
			fmt.Print("\t", floor, "\t")
			if all_upward_requests[responsible_id][floor] != zero_request {fmt.Print("*")
			}else {fmt.Print(" ")}
			fmt.Print("\t")
			if all_downward_requests[responsible_id][floor] != zero_request {fmt.Print("*")
			}else {fmt.Print(" ")}
			fmt.Print("\t")
			if all_command_requests[responsible_id][floor] != zero_request {fmt.Print("*")
			}else {fmt.Print(" ")}
			fmt.Print("\n")
		}
		fmt.Println("--------------------------------------")
		fmt.Print("\n")
	}
	fmt.Print("\n\n\n\n")

}