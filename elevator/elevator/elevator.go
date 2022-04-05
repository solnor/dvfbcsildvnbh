package elevator

import (
	"driver/cost"
	"elevator/config"
	"elevator/elevio"
	fsm "elevator/fsm"
	"elevator/timer"
	"fmt"
)

// fsm "elevator/fsm"

func Elevator_Run() {
	elevio.Init("localhost:15657", config.NumFloors)
	fsm.Fsm_init()

	var doorTimeout = make(chan bool)
	drv_buttons := make(chan config.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.Observer(doorTimeout)

	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(config.MD_Down)
		fsm.Elevator1.Dirn = config.MD_Down
		fsm.Elevator1.Behaviour = config.EB_Moving
	}

	for {
		select {
		case a := <-drv_buttons:
			fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
			fmt.Printf("Cost: %d", cost.TimeToIdle(fsm.Elevator1))
		case a := <-drv_floors:
			fsm.Fsm_onFloorArrival(a)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				fsm.Elevator1.Obstruction = true
			} else {
				fsm.Elevator1.Obstruction = false
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < config.NumFloors; f++ {
				for b := config.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		case <-doorTimeout:
			fsm.Fsm_onDoorTimeout()
		}
	}
}

// func Elevator_Run() {
// 	elevio.Init("localhost:15657", config.NumFloors)
// 	fsm.Fsm_init()

// 	var doorTimeout = make(chan bool)
// 	drv_buttons := make(chan config.ButtonEvent)
// 	drv_floors := make(chan int)
// 	drv_obstr := make(chan bool)
// 	drv_stop := make(chan bool)

// 	go elevio.PollButtons(drv_buttons)
// 	go elevio.PollFloorSensor(drv_floors)
// 	go elevio.PollObstructionSwitch(drv_obstr)
// 	go elevio.PollStopButton(drv_stop)
// 	go timer.Observer(doorTimeout)

// 	if elevio.GetFloor() == -1 {
// 		elevio.SetMotorDirection(config.MD_Down)
// 		fsm.Elevator1.Dirn = config.MD_Down
// 		fsm.Elevator1.Behaviour = config.EB_Moving
// 	}

// 	for {
// 		select {
// 		case a := <-drv_buttons:
// 			fsm.Fsm_onRequestButtonPress(a.Floor, a.Button)
// 		case a := <-drv_floors:
// 			fsm.Fsm_onFloorArrival(a)

// 		case a := <-drv_obstr:
// 			fmt.Printf("%+v\n", a)
// 			if a {
// 				fsm.Elevator1.Obstruction = true
// 			} else {
// 				fsm.Elevator1.Obstruction = false
// 			}

// 		case a := <-drv_stop:
// 			fmt.Printf("%+v\n", a)
// 			for f := 0; f < config.NumFloors; f++ {
// 				for b := config.ButtonType(0); b < 3; b++ {
// 					elevio.SetButtonLamp(b, f, false)
// 				}
// 			}
// 		case <-doorTimeout:
// 			fsm.Fsm_onDoorTimeout()
// 		}
// 	}
// }

// func eb_toString(eb config.ElevatorBehaviour) string {
// 	if eb == config.EB_Idle {
// 		return "EB_idle"
// 	} else if eb == config.EB_DoorOpen {
// 		return "EB_DoorOpen"
// 	} else if eb == config.EB_Moving {
// 		return "EB_Moving"
// 	} else {
// 		return "EB_UNDEFINED"
// 	}
// 	// eb == EB_Idle       ? "EB_Idle"         :
// 	// eb == EB_DoorOpen   ? "EB_DoorOpen"     :
// 	// eb == EB_Moving     ? "EB_Moving"       :
// 	//                       "EB_UNDEFINED"    ;
// }

// func Elevator_print(es config.Elevator) {
// 	fmt.Printf("  +--------------------+\n")
// 	fmt.Printf("  |floor = %-2d          |\n  |dirn  = %-12.12s|\n  |behav = %-12.12s|\n",
// 		es.Floor,
// 		elevio_dirn_toString(es.Dirn),
// 		eb_toString(es.Behaviour))
// 	fmt.Printf("  +--------------------+\n")
// 	fmt.Printf("  |  | up  | dn  | cab |\n")
// 	for f := config.N_FLOORS - 1; f >= 0; f-- {
// 		fmt.Printf("  | %d", f)
// 		var btn config.ButtonType
// 		for btn = 0; btn <= config.BT_Cab; btn++ {
// 			if (f == config.N_FLOORS-1 && btn == config.BT_HallUp) ||
// 				(f == 0 && btn == config.BT_HallDown) {
// 				fmt.Printf("|     ")
// 			} else {
// 				if es.Requests[f][btn] == 1 {
// 					fmt.Printf("|  #  ")
// 				} else {
// 					fmt.Printf("|  -  ")
// 				}
// 				// fmt.Printf(es.requests[f][btn] ? "|  #  " : "|  -  ");
// 			}
// 	fmt.Printf("  +--------------------+\n")
// }

// func elevio_dirn_toString(d config.MotorDirection) string {
// 	if d == config.MD_Up {
// 		return "MD_Up"
// 	} else if d == config.MD_Down {
// 		return "MD_Down"
// 	} else if d == config.MD_Stop {
// 		return "MD_Stop"
// 	} else {
// 		return "MD_UNDEFINED"
// 	}

// 	// return
// 	//     d == D_Up    ? "D_Up"         :
// 	//     d == D_Down  ? "D_Down"       :
// 	//     d == D_Stop  ? "D_Stop"       :
// 	//                    "D_UNDEFINED"  ;
// }

// func elevio_button_toString(b config.ButtonType) string {
// 	if b == config.BT_HallUp {
// 		return "B_HallUp"
// 	} else if b == config.BT_HallUp {
// 		return "B_HallDown"
// 	} else if b == config.BT_HallUp {
// 		return "B_Cab"
// 	} else {
// 		return "B_UNDEFINED"
// 	}

// 	// return
// 	//     b == B_HallUp       ? "B_HallUp"        :
// 	//     b == B_HallDown     ? "B_HallDown"      :
// 	//     b == B_Cab          ? "B_Cab"           :
// 	//                           "B_UNDEFINED"     ;
// }
