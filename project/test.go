package test//main

import(
	"./request_watchdog"
	"./message_structs"
	"./hardware_interface"
	"time"
	"fmt"
)

func main() {
	timed_out_request_chan := make(chan message_structs.Request, 1)
	request_watchdog.Init(timed_out_request_chan)

	id := "192.168.1.test"
	request := message_structs.Request{
		Floor: 0, 
		Message_origin_id: id, 
		Request_type: hardware_interface.BUTTON_TYPE_COMMAND, 
		Responsible_elevator: id, 
		Is_completed: false}
	request_watchdog.Request_timer_start(request)

	select{
		case timed_out_request := <-timed_out_request_chan:
			fmt.Println("Request timed out: ", timed_out_request)
		case <-time.After(time.Second * 10):
			request_watchdog.Request_timer_stop(request)
			fmt.Println("No timed out request. ")
	}
}