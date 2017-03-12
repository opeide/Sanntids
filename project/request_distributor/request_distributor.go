package request_distributor

import (
	"../hardware_interface"
	"../message_structs"
	"../network/peers"
	"../request_executor"
	"../request_watchdog"
	"fmt"
	"os"
	"sort"
	"time"
)

var all_elevator_states = make(map[string]message_structs.Elevator_state)
var all_upward_requests = make(map[string][]message_structs.Request)
var all_downward_requests = make(map[string][]message_structs.Request)
var all_command_requests = make(map[string][]message_structs.Request)

var zero_request message_structs.Request = message_structs.Request{}

var local_id string
var network_request_tx_chan chan<- message_structs.Request
var local_elevator_state_changes_tx_chan chan<- message_structs.Elevator_state
var requests_to_execute_chan chan<- message_structs.Request
var set_lamp_chan chan<- message_structs.Set_lamp_message

const (
	num_network_transmit_repeats = 3
	network_transmit_wait_time   = 5 // Milliseconds
)

// [solved] Remember to do caborders when we get back after losing power!!!!!!!!!!!!!!! solved: get them from network
// If we lose network, will we be filling up the network channel and eventually block???????

func Distribute_requests(
	local_id_parameter string,
	network_request_tx_chan_parameter chan<- message_structs.Request,
	local_elevator_state_changes_tx_chan_parameter chan<- message_structs.Elevator_state,
	requests_to_execute_chan_parameter chan<- message_structs.Request,
	set_lamp_chan_parameter chan<- message_structs.Set_lamp_message,
	peer_update_chan <-chan peers.PeerUpdate,
	network_request_rx_chan <-chan message_structs.Request,
	non_local_elevator_state_changes_rx_chan <-chan message_structs.Elevator_state,
	button_request_chan <-chan message_structs.Request,
	executed_requests_chan <-chan message_structs.Request,
	local_elevator_state_changes_chan <-chan message_structs.Elevator_state,
	timed_out_requests_chan <-chan message_structs.Request) {

	local_id = local_id_parameter
	network_request_tx_chan = network_request_tx_chan_parameter
	local_elevator_state_changes_tx_chan = local_elevator_state_changes_tx_chan_parameter
	requests_to_execute_chan = requests_to_execute_chan_parameter
	set_lamp_chan = set_lamp_chan_parameter

	select {
	case local_elevator_state := <-local_elevator_state_changes_chan:
		update_own_state_and_send_to_network(local_elevator_state)

		all_upward_requests[local_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_downward_requests[local_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_command_requests[local_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
	}

	for {
		select {
		case peer_update := <-peer_update_chan:
			if peer_update.New != "" {
				fmt.Println("Distributor: New Peer: ", peer_update.New)
				if _, ok := all_command_requests[peer_update.New]; !ok {
					fmt.Println("creating new request list for brand new peer")
					all_upward_requests[peer_update.New] = make([]message_structs.Request, hardware_interface.N_FLOORS)
					all_downward_requests[peer_update.New] = make([]message_structs.Request, hardware_interface.N_FLOORS)
					all_command_requests[peer_update.New] = make([]message_structs.Request, hardware_interface.N_FLOORS)
				}

				if peer_update.New != local_id {
					update_own_state_and_send_to_network(all_elevator_states[local_id])
				}

				// new peer, send all stored requests on network
				for _, requests_list_by_id := range []map[string][]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
					for _, request_list_by_floor := range requests_list_by_id {
						for _, request := range request_list_by_floor {
							if request != zero_request {
								distribute_request(request)
							}
						}
					}
				}
			}

			if len(peer_update.Lost) != 0 {
				for _, lost_elevator_id := range peer_update.Lost {
					fmt.Println("Lost elevator: ", lost_elevator_id)
					if lost_elevator_id != local_id {
						fmt.Println("Lost non-local elevator: ", lost_elevator_id)
						//delete elevator and Inherit requests
						delete(all_elevator_states, lost_elevator_id)
						for _, requests_list_by_id := range []map[string][]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
							for responsible_id, request_list_by_floor := range requests_list_by_id {
								if responsible_id == lost_elevator_id {
									for _, request := range request_list_by_floor {
										if request != zero_request {
											delete_request_except_non_completed_command(request)
											if request.Request_type != hardware_interface.BUTTON_TYPE_COMMAND {
												request.Responsible_elevator = local_id
											}
											if request == zero_request { // TEMP TEST
												fmt.Println("Peer lost, trying to register zero request. ")
											}
											register_request(request)
											distribute_request(request)
										}
									}
									break
								}
							}
						}
					}
				}
			}

		case local_elevator_state := <-local_elevator_state_changes_chan:
			update_own_state_and_send_to_network(local_elevator_state)

		case button_request := <-button_request_chan:
			button_request.Responsible_elevator = decide_responsible_elevator(button_request)

			if button_request == zero_request { // TEMP TEST
				fmt.Println("Local request, trying to register zero request. ")
			}
			register_request(button_request)
			distribute_request(button_request)

		case executed_request := <-executed_requests_chan:
			distribute_request(executed_request)
			register_finished_request(executed_request)

		case non_local_request := <-network_request_rx_chan:
			if non_local_request.Message_origin_id == local_id {
				break
			}
			if non_local_request.Is_completed {
				register_finished_request(non_local_request)
			} else {
				if non_local_request == zero_request { // TEMP TEST
					fmt.Println("Non_local_request, trying to register zero request. ")
				}
				register_request(non_local_request)
				if non_local_request.Responsible_elevator == local_id {
					distribute_request(non_local_request)
				}
			}

		case non_local_elevator_state := <-non_local_elevator_state_changes_rx_chan:
			if non_local_elevator_state.Elevator_id == local_id {
				break
			}
			if non_local_elevator_state.Elevator_id == "" { // TEMP TEST
				fmt.Println("Got non local elevator state with no id")
			}
			all_elevator_states[non_local_elevator_state.Elevator_id] = non_local_elevator_state

		case timed_out_request := <-timed_out_requests_chan:
			if timed_out_request == zero_request { // TEMP TEST
				fmt.Println("Timed out request, trying to register zero request. ")
			}
			if timed_out_request.Responsible_elevator == local_id {
				fmt.Println("Timed out on local request: ", timed_out_request)
				register_request(timed_out_request)
				distribute_request(timed_out_request)
				os.Exit(0) //Lets backup take over (effectively a program restart)
				break
			}
			fmt.Println("Timed out on non-local request: ", timed_out_request)
			timed_out_request.Responsible_elevator = local_id
			register_request(timed_out_request)
			distribute_request(timed_out_request)
		}
	}
}

func register_request(request message_structs.Request) {
	if request.Responsible_elevator == "" {
		fmt.Println("Trying to register blank id! ERROR")
	}
	if _, ok := all_command_requests[request.Responsible_elevator]; !ok {
		fmt.Println("saved register!")
		all_upward_requests[request.Responsible_elevator] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_downward_requests[request.Responsible_elevator] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_command_requests[request.Responsible_elevator] = make([]message_structs.Request, hardware_interface.N_FLOORS)
	}
	request_watchdog.Timer_start(request)
	switch request.Request_type {
	case hardware_interface.BUTTON_TYPE_CALL_UP:
		if all_upward_requests[request.Responsible_elevator][request.Floor] == zero_request {
			all_upward_requests[request.Responsible_elevator][request.Floor] = request
		}
	case hardware_interface.BUTTON_TYPE_CALL_DOWN:
		if all_downward_requests[request.Responsible_elevator][request.Floor] == zero_request {
			all_downward_requests[request.Responsible_elevator][request.Floor] = request
		}
	case hardware_interface.BUTTON_TYPE_COMMAND:
		if all_command_requests[request.Responsible_elevator][request.Floor] == zero_request {
			all_command_requests[request.Responsible_elevator][request.Floor] = request
		}
	}
	set_request_lights(request, 1)
	print_request_list()
}

func distribute_request(request message_structs.Request) {
	if request.Responsible_elevator == local_id && request.Is_completed == false {
		requests_to_execute_chan <- request
	}
	request.Message_origin_id = local_id

	for sendt_n_times := 1; sendt_n_times <= num_network_transmit_repeats; sendt_n_times++ {
		network_request_tx_chan <- request
		<-time.After(time.Millisecond * network_transmit_wait_time)
	}
}

func update_own_state_and_send_to_network(elevator_state message_structs.Elevator_state) {
	elevator_state.Elevator_id = local_id
	all_elevator_states[local_id] = elevator_state

	for sendt_n_times := 1; sendt_n_times <= num_network_transmit_repeats; sendt_n_times++ {
		local_elevator_state_changes_tx_chan <- elevator_state
		<-time.After(time.Millisecond * network_transmit_wait_time)
	}
}

func register_finished_request(request message_structs.Request) {
	set_request_lights(request, 0)
	delete_request_except_non_completed_command(request)
	print_request_list()
}

func delete_request_except_non_completed_command(request message_structs.Request) {
	all_upward_requests[request.Responsible_elevator][request.Floor] = zero_request
	all_downward_requests[request.Responsible_elevator][request.Floor] = zero_request
	if request.Is_completed {
		all_command_requests[request.Responsible_elevator][request.Floor] = zero_request
	}
	request_watchdog.Timer_stop(request)
}

func set_request_lights(request message_structs.Request, value int) {
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
			if request.Responsible_elevator == local_id {
				set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
				set_lamp_chan <- set_lamp_message
			}
		}
	} else {
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_UP
		set_lamp_chan <- set_lamp_message
		set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_DOWN
		set_lamp_chan <- set_lamp_message

		if request.Responsible_elevator == local_id {
			set_lamp_message.Lamp_type = hardware_interface.LAMP_TYPE_COMMAND
			set_lamp_chan <- set_lamp_message
		}
	}
}

func decide_responsible_elevator(request message_structs.Request) string {
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
	fmt.Println(fastest_elevator_id, " would solve request fastest: ", fastest_time, request)
	return fastest_elevator_id
}

func estimate_time_for_elevator_to_complete_request(elevator_state message_structs.Elevator_state, request message_structs.Request) int {
	elevator_id := elevator_state.Elevator_id

	if _, ok := all_command_requests[elevator_id]; !ok {
		fmt.Println("Saved estimate!")
		all_upward_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_downward_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_command_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
	}

	request_list_by_motor_direction := map[int]map[string][]message_structs.Request{hardware_interface.MOTOR_DIRECTION_DOWN: all_downward_requests, hardware_interface.MOTOR_DIRECTION_UP: all_upward_requests}
	endfloor_in_motor_direction := map[int]int{hardware_interface.MOTOR_DIRECTION_DOWN: -1, hardware_interface.MOTOR_DIRECTION_UP: hardware_interface.N_FLOORS}

	total_time := 0
	starting_floor := elevator_state.Last_visited_floor
	last_stop_floor := starting_floor
	starting_motor_direction := elevator_state.Last_non_stop_motor_direction

	for _, motor_direction := range []int{starting_motor_direction, -starting_motor_direction, starting_motor_direction} { // Using that motor directions are +/-1. See hardware_interface for definitions
		for floor := starting_floor; floor != endfloor_in_motor_direction[motor_direction]; floor += motor_direction {
			if floor == request.Floor {
				total_time += abs(last_stop_floor-floor) * request_executor.ESTIMATED_TIME_BETWEEN_FLOORS
				last_stop_floor = floor
				if request.Request_type == hardware_interface.BUTTON_TYPE_COMMAND ||
					request.Request_type == hardware_interface.BUTTON_TYPE_CALL_UP && motor_direction == hardware_interface.MOTOR_DIRECTION_UP ||
					request.Request_type == hardware_interface.BUTTON_TYPE_CALL_DOWN && motor_direction == hardware_interface.MOTOR_DIRECTION_DOWN {

					return total_time
				}
			}

			if request_list_by_motor_direction[motor_direction][elevator_id][floor] != zero_request || all_command_requests[elevator_id][floor] != zero_request {
				total_time += abs(last_stop_floor-floor) * request_executor.ESTIMATED_TIME_BETWEEN_FLOORS
				if request.Floor == floor {
					return total_time
				}
				total_time += request_executor.DOOR_OPEN_TIME
				last_stop_floor = floor
			}
		}
		starting_floor = last_stop_floor
	}

	fmt.Println("SIMULATOR NEVER REACHED FLOOR! SHOULD NOT HAPPEN!")
	return total_time
}

func abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

func print_request_list() {
	var sorted_ids []string
	for id := range all_command_requests {
		sorted_ids = append(sorted_ids, id)
	}
	sort.Strings(sorted_ids)

	fmt.Print("\n\n\n\n")
	for _, responsible_id := range sorted_ids {
		fmt.Print("\n")
		fmt.Println("Responsible: ", responsible_id)
		fmt.Println("--------------------------------------")
		fmt.Println("\tFLOOR\tUP\tDOWN\tCOMMAND")
		for floor := hardware_interface.N_FLOORS - 1; floor >= 0; floor-- {
			fmt.Print("\t", floor, "\t")
			if all_upward_requests[responsible_id][floor] != zero_request {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
			fmt.Print("\t")
			if all_downward_requests[responsible_id][floor] != zero_request {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
			fmt.Print("\t")
			if all_command_requests[responsible_id][floor] != zero_request {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
			fmt.Print("\t")
			if all_elevator_states[responsible_id].Last_visited_floor == floor {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
			fmt.Print("\n")
		}
		fmt.Println("--------------------------------------")
		fmt.Print("\n")
	}
	fmt.Print("\n\n\n\n")

}
