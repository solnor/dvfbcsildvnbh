package requests

import (
	"elevator/config"
	. "elevator/config"
	. "elevator/elevio"
	//"golang.org/x/text/cases"
	//"fmt"
)

var Obstruction bool = false

func Requests_above(e Elevator) int {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] == 1 {
				return 1
			}
		}
	}
	return 0
}

func Requests_below(e Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] == 1 { //should be changed?
				return 1
			}
		}

	}
	return 0
}

func Requests_here(e Elevator) int {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] == 1 {
			return 1
		}
	}
	return 0
}

func Requests_nextAction(e Elevator) config.Action {
	var a config.Action
	switch e.Dirn {
	case MD_Up:

		if Requests_above(e) == 1 {
			a.Dirn = MD_Up
			a.Behaviour = EB_Moving
		} else if Requests_below(e) == 1 {
			a.Dirn = MD_Down
			a.Behaviour = EB_Moving
		} else if Requests_here(e) == 1 {
			a.Dirn = MD_Down
			a.Behaviour = EB_DoorOpen
		} else {
			a.Dirn = MD_Stop
			a.Behaviour = EB_Idle
		}
		break

	case MD_Down:
		if Requests_above(e) == 1 {
			a.Dirn = MD_Up
			a.Behaviour = EB_Moving
		} else if Requests_below(e) == 1 {
			a.Dirn = MD_Down
			a.Behaviour = EB_Moving
		} else if Requests_here(e) == 1 {
			a.Dirn = MD_Up
			a.Behaviour = EB_DoorOpen
		} else {
			a.Dirn = MD_Stop
			a.Behaviour = EB_Idle
		}
		break

	case MD_Stop:
		if e.Obstruction {
			a.Dirn = MD_Stop
			a.Behaviour = EB_DoorOpen
		} else {
			if Requests_above(e) == 1 {
				a.Dirn = MD_Up
				a.Behaviour = EB_Moving
			} else if Requests_below(e) == 1 {
				a.Dirn = MD_Down
				a.Behaviour = EB_Moving
			} else if Requests_here(e) == 1 {
				a.Dirn = MD_Stop
				a.Behaviour = EB_DoorOpen
			} else {
				a.Dirn = MD_Stop
				a.Behaviour = EB_Idle
			}
		}
		break

	default:
		a.Dirn = MD_Stop
		a.Behaviour = EB_Idle
		break
	}
	// if Obstruction {
	// 	a.Behaviour = EB_DoorOpen
	// }
	return a
}

func Requests_shouldStop(e Elevator) bool {
	var x bool
	switch e.Dirn {
	case MD_Down:
		x = ToBool(byte(e.Requests[e.Floor][BT_HallDown])) ||
			ToBool(byte(e.Requests[e.Floor][BT_Cab])) ||
			!(ToBool(byte(Requests_below(e))))
		//return x

	case MD_Up:
		x = ToBool(byte(e.Requests[e.Floor][BT_HallUp])) ||
			ToBool(byte(e.Requests[e.Floor][BT_Cab])) ||
			!(ToBool(byte(Requests_above(e))))
		//return x

	case MD_Stop:
		x = true //-------------------------------------------------------Added due to cost func
	default:
		x = false
	}

	return x
}

func Requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		return e.Floor == btn_floor

	case CV_InDirn:
		return e.Floor == btn_floor && ((e.Dirn == MD_Up && btn_type == BT_HallUp) ||
			(e.Dirn == MD_Down && btn_type == BT_HallDown) ||
			e.Dirn == MD_Stop ||
			btn_type == BT_Cab)

	default:
		return false
	}
}

func Requests_clearAtCurrentFloor(e config.Elevator, onClearedRequest func(e config.Elevator, btn config.ButtonType, floor int)) Elevator {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < N_BUTTONS; btn++ {
			// if onClearedRequest != nil {
			// 	onClearedRequest(e, )
			// }
			e.Requests[e.Floor][btn] = 0
		}

	case CV_InDirn:
		e.Requests[e.Floor][BT_Cab] = 0
		switch e.Dirn {
		case MD_Up:
			if !ToBool(byte(Requests_above(e))) && !ToBool(byte(e.Requests[e.Floor][BT_HallUp])) {
				e.Requests[e.Floor][BT_HallDown] = 0
			}
			e.Requests[e.Floor][BT_HallUp] = 0

		case MD_Down:
			if !ToBool(byte(Requests_below(e))) && !ToBool(byte(e.Requests[e.Floor][BT_HallDown])) {
				e.Requests[e.Floor][BT_HallUp] = 0
			}
			e.Requests[e.Floor][BT_HallDown] = 0

		case MD_Stop:

		default:
			e.Requests[e.Floor][BT_HallUp] = 0
			e.Requests[e.Floor][BT_HallDown] = 0

		}

	default:
		break
	}
	return e
}

// type Request struct {
// 	dirn  config.MotorDirection
// 	floor int
// }

func Requests_onClearedRequest(e config.Elevator, btn config.ButtonType, floor int) {
	e.Requests[floor][btn] = 0
}
