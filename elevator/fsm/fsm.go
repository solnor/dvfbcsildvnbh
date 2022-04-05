package fsm

import (
	config "elevator/config"
	//. "elevator/elevUI"
	"reflect"

	// . "elevator/elevator"
	. "elevator/elevio"
	. "elevator/requests"
	timer "elevator/timer"
	"fmt"
	"runtime"
	"time"
)

var ThisElevator config.Elevator

func Fsm_init() {
	ThisElevator = config.NewElevator()
	ThisElevator.Floor = -1
	ThisElevator.Dirn = config.MD_Stop
	ThisElevator.Behaviour = config.EB_Idle
	ThisElevator.Obstruction = false

	ThisElevator.Config.ClearRequestVariant = config.CV_All
	ThisElevator.Config.DoorOpenDuration_s = 3
}

// var DoorTimer = time.NewTimer(time.Duration(3 * time.Second))

var PrevDir config.MotorDirection

func GetFunctionname(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func SetAllLights(es config.Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		//for brn := range ButtonType { //:= 0; btn < N_BUTTONS; btn++ {
		var btn config.ButtonType
		for btn = config.BT_HallUp; btn <= config.BT_Cab; btn++ {
			if es.Requests[floor][btn] == 1 {
				SetButtonLamp(btn, floor, true)
			} else {
				SetButtonLamp(btn, floor, false)
			}
		}

	}
}

func SetCabLights(es config.Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		if es.Requests[floor][2] == 1 {
			SetButtonLamp(2, floor, true)
		} else {
			SetButtonLamp(2, floor, false)
		}

	}
}

func Fsm_onRequestButtonPress(btn_floor int, btn_type config.ButtonType) {
	fmt.Printf("\n\n%s(%d, %v)\n", GetFunctionname(Fsm_onRequestButtonPress), btn_floor, btn_type)
	//Elevator_print(Elevator1)

	switch ThisElevator.Behaviour {
	case config.EB_DoorOpen:
		if Requests_shouldClearImmediately(ThisElevator, btn_floor, btn_type) {
			//timer_start.reset(elevator.Config.DoorOpenDuration_s)
			timer.Reset(time.Duration(config.DOOR_OPEN_TIME_S) * time.Second)
		} else {
			ThisElevator.Requests[btn_floor][btn_type] = 1
		}
		break
	case config.EB_Moving:
		ThisElevator.Requests[btn_floor][btn_type] = 1
		break
	case config.EB_Idle:
		ThisElevator.Requests[btn_floor][btn_type] = 1
		var a config.Action
		a = Requests_nextAction(ThisElevator)
		ThisElevator.Dirn = a.Dirn
		ThisElevator.Behaviour = a.Behaviour
		switch a.Behaviour {
		case config.EB_DoorOpen:
			SetDoorOpenLamp(true)
			timer.Reset(time.Duration(config.DOOR_OPEN_TIME_S) * time.Second)
			ThisElevator = Requests_clearAtCurrentFloor(ThisElevator, Requests_onClearedRequest)
			break
		case config.EB_Moving:
			// if !Obstruction
			SetMotorDirection(ThisElevator.Dirn)
			// fmt.Printf("Set motor direction to %v\n", Elevator1.Dirn)
			break
		case config.EB_Idle:
			break
		}
		break
	}

	// setAllLights(ThisElevator) //// Commented out due to button light contract
	SetCabLights(ThisElevator)
	//fmt.Printf("\nNew State: \n")
	//Elevator_print(Elevator1)
}

func Fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\n%s(%d)\n\n", GetFunctionname(Fsm_onFloorArrival), newFloor)
	//Elevator_print(Elevator1)
	ThisElevator.Floor = newFloor
	SetFloorIndicator(ThisElevator.Floor)

	switch ThisElevator.Behaviour {
	case config.EB_Moving:
		if Requests_shouldStop(ThisElevator) {
			ThisElevator.Dirn = config.MD_Stop
			SetMotorDirection(ThisElevator.Dirn)
			SetDoorOpenLamp(true)
			ThisElevator = Requests_clearAtCurrentFloor(ThisElevator, Requests_onClearedRequest)
			//timer_start(elevator.Config.DoorOpenDuration_s)
			timer.Reset(time.Duration(config.DOOR_OPEN_TIME_S) * time.Second)
			SetAllLights(ThisElevator)
			ThisElevator.Behaviour = config.EB_DoorOpen
		}
		break
	default:
		break
	}
	//fmt.Printf("\nNew State: \n")
	//Elevator_print(Elevator1)
}

func Fsm_onDoorTimeout() {
	fmt.Printf("\n\n%s()\n\n", GetFunctionname(Fsm_onDoorTimeout))
	//Elevator_print(Elevator1)

	switch ThisElevator.Behaviour {
	case config.EB_DoorOpen:
		// if Obstruction {
		// 	break
		// }
		var a config.Action
		a = Requests_nextAction(ThisElevator)
		ThisElevator.Dirn = a.Dirn
		ThisElevator.Behaviour = a.Behaviour
		switch ThisElevator.Behaviour {
		case config.EB_DoorOpen:
			//timer_start(elevator.Config.DoorOpenDuration_s)
			timer.Reset(time.Duration(config.DOOR_OPEN_TIME_S) * time.Second)

			ThisElevator = Requests_clearAtCurrentFloor(ThisElevator, Requests_onClearedRequest)
			SetAllLights(ThisElevator)
			break
		case config.EB_Moving:
			SetDoorOpenLamp(false)
			SetMotorDirection(ThisElevator.Dirn)
		case config.EB_Idle:
			SetDoorOpenLamp(false)
			SetMotorDirection(ThisElevator.Dirn)
			break
		}

		break
	default:
		break
	}

	//fmt.Printf("\nNew state: \n")
	//Elevator_print(Elevator1)
}
