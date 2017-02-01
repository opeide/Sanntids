package main

import (
	"./network_request_distributor"
	//"fmt"
)

func main() {
	go network_request_distributor.Start()
}
