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
	// driver.Drive()
	driver.Elevator_Run()
	// elevio.Init("localhost:15657", config.NumFloors)
	// fsm.Fsm_init()

	// var doorTimeout = make(chan bool)
	// drv_buttons := make(chan config.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)
	// go timer.Observer(doorTimeout)

	// if elevio.GetFloor() == -1 {
	// 	elevio.SetMotorDirection(config.MD_Down)
	// 	fsm.Elevator1.Dirn = config.MD_Down
	// 	fsm.Elevator1.Behaviour = config.EB_Moving
	// }

	// for {
	// 	select {
	// 	case a := <-drv_buttons:
	// 		fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
	// 	case a := <-drv_floors:
	// 		fsm.Fsm_onFloorArrival(a)

	// 	case a := <-drv_obstr:
	// 		fmt.Printf("%+v\n", a)
	// 		if a {
	// 			fsm.Elevator1.Obstruction = true
	// 		} else {
	// 			fsm.Elevator1.Obstruction = false
	// 		}

	// 	case a := <-drv_stop:
	// 		fmt.Printf("%+v\n", a)
	// 		for f := 0; f < config.NumFloors; f++ {
	// 			for b := config.ButtonType(0); b < 3; b++ {
	// 				elevio.SetButtonLamp(b, f, false)
	// 			}
	// 		}
	// 	case <-doorTimeout:
	// 		fsm.Fsm_onDoorTimeout()
	// 	}
	// }
}
