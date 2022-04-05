package main

import (
	"driver/driver"
	"fmt"
	// "elevator/config"
	// "elevator/elevio"
	// // . "elevator/requests"
	// fsm "elevator/fsm"
	// timer "elevator/timer"
	// . "../Driver-go/timer"
)

func main() {
	fmt.Printf("Startup\n")
	driver.Elevator_Run()
}
