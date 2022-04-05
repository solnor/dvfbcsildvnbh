package elevUI

import (
	"elevator/config"
	"fmt"
)

func eb_toString(eb config.ElevatorBehaviour) string {
	if eb == config.EB_Idle {
		return "EB_idle"
	} else if eb == config.EB_DoorOpen {
		return "EB_DoorOpen"
	} else if eb == config.EB_Moving {
		return "EB_Moving"
	} else {
		return "EB_UNDEFINED"
	}
}

func Elevator_print(es config.Elevator) {
	fmt.Printf("  +--------------------+\n")
	fmt.Printf("  |floor = %-2d          |\n  |dirn  = %-12.12s|\n  |behav = %-12.12s|\n",
		es.Floor,
		elevio_dirn_toString(es.Dirn),
		eb_toString(es.Behaviour))
	fmt.Printf("  +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := config.N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		var btn config.ButtonType
		for btn = 0; btn <= config.BT_Cab; btn++ {
			if (f == config.N_FLOORS-1 && btn == config.BT_HallUp) ||
				(f == 0 && btn == config.BT_HallDown) {
				fmt.Printf("|     ")
			} else {
				if es.Requests[f][btn] == 1 {
					fmt.Printf("|  #  ")
				} else {
					fmt.Printf("|  -  ")
				}
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("  +--------------------+\n")
}

func elevio_dirn_toString(d config.MotorDirection) string {
	if d == config.MD_Up {
		return "MD_Up"
	} else if d == config.MD_Down {
		return "MD_Down"
	} else if d == config.MD_Stop {
		return "MD_Stop"
	} else {
		return "MD_UNDEFINED"
	}
}

func elevio_button_toString(b config.ButtonType) string {
	if b == config.BT_HallUp {
		return "B_HallUp"
	} else if b == config.BT_HallUp {
		return "B_HallDown"
	} else if b == config.BT_HallUp {
		return "B_Cab"
	} else {
		return "B_UNDEFINED"
	}
}
