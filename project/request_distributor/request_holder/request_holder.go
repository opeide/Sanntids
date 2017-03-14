package request_holder

import (
	"../../hardware_interface"
	"../../message_structs"
	"fmt"
	"sort"
)

var all_upward_requests = make(map[string][]message_structs.Request)
var all_downward_requests = make(map[string][]message_structs.Request)
var all_command_requests = make(map[string][]message_structs.Request)

var zero_request message_structs.Request = message_structs.Request{}

func Attempt_init_id(elevator_id string) {
	attempt_init_request_lists_for_id(elevator_id)
}

func attempt_init_request_lists_for_id(elevator_id string) {
	if _, ok := all_command_requests[elevator_id]; !ok {
		all_upward_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_downward_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
		all_command_requests[elevator_id] = make([]message_structs.Request, hardware_interface.N_FLOORS)
	}
}

func Hold_request(request message_structs.Request) {
	if request.Responsible_elevator == "" {
		return
	}
	attempt_init_request_lists_for_id(request.Responsible_elevator)

	switch request.Request_type {
	case hardware_interface.BUTTON_TYPE_CALL_UP:
		if all_upward_requests[request.Responsible_elevator][request.Floor] == zero_request { //in holder
			all_upward_requests[request.Responsible_elevator][request.Floor] = request
		}
	case hardware_interface.BUTTON_TYPE_CALL_DOWN:
		if all_downward_requests[request.Responsible_elevator][request.Floor] == zero_request { //
			all_downward_requests[request.Responsible_elevator][request.Floor] = request
		}
	case hardware_interface.BUTTON_TYPE_COMMAND:
		if all_command_requests[request.Responsible_elevator][request.Floor] == zero_request { //
			all_command_requests[request.Responsible_elevator][request.Floor] = request
		}
	}
}

func Delete_single_request(request message_structs.Request) {
	switch request.Request_type {
	case hardware_interface.BUTTON_TYPE_CALL_DOWN:
		all_downward_requests[request.Responsible_elevator][request.Floor] = zero_request
	case hardware_interface.BUTTON_TYPE_CALL_UP:
		all_upward_requests[request.Responsible_elevator][request.Floor] = zero_request
	case hardware_interface.BUTTON_TYPE_COMMAND:
		all_command_requests[request.Responsible_elevator][request.Floor] = zero_request
	}
}

func Delete_floor_requests_not_uncompleted_command(request message_structs.Request) {
	for _, elevator_id := all_command_requests{
		all_upward_requests[elevator_id][request.Floor] = zero_request
		all_downward_requests[elevator_id][request.Floor] = zero_request
	}
	if request.Is_completed {
		all_command_requests[request.Responsible_elevator][request.Floor] = zero_request
	}
}

func Get_all_requests() []message_structs.Request {
	all_requests := make([]message_structs.Request, 0)

	for _, requests_list_by_id := range []map[string][]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
		for _, requests_by_floor := range requests_list_by_id {
			for _, request := range requests_by_floor {
				all_requests = append(all_requests, request)
			}
		}
	}
	return all_requests
}

func Get_all_requests_for_id(elevator_id string) []message_structs.Request {
	attempt_init_request_lists_for_id(elevator_id)

	all_requests_for_id := make([]message_structs.Request, 0)

	for _, requests_list_for_id := range []map[string][]message_structs.Request{all_upward_requests, all_downward_requests, all_command_requests} {
		for _, request := range requests_list_for_id[elevator_id] {
			if request != zero_request {
				all_requests_for_id = append(all_requests_for_id, request)
			}
		}
	}
	return all_requests_for_id
}

func Get_requests_for_motor_direction(motor_direction int) map[string][]message_structs.Request {
	switch motor_direction {
	case hardware_interface.MOTOR_DIRECTION_UP:
		return all_upward_requests
	case hardware_interface.MOTOR_DIRECTION_DOWN:
		return all_downward_requests
	case hardware_interface.MOTOR_DIRECTION_STOP:
		return all_downward_requests
	}
	return make(map[string][]message_structs.Request)
}

func Get_requests_of_type(request_type int) map[string][]message_structs.Request {
	switch request_type {
	case hardware_interface.BUTTON_TYPE_CALL_UP:
		return all_upward_requests
	case hardware_interface.BUTTON_TYPE_CALL_DOWN:
		return all_downward_requests
	case hardware_interface.BUTTON_TYPE_COMMAND:
		return all_downward_requests
	}
	return make(map[string][]message_structs.Request)
}

func Print_requests(all_elevator_states map[string]message_structs.Elevator_state) {
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
