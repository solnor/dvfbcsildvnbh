package cost

import (
	"elevator/config"
	"elevator/requests"
)

func TimeToIdle(e config.Elevator) int64 {
	testElev := config.DupElevator(e)
	var duration int64 = 0
	var a config.Action

	switch testElev.Behaviour {
	case config.EB_Idle:
		a = requests.Requests_nextAction(testElev)
		testElev.Dirn = a.Dirn
		if testElev.Dirn == config.MD_Stop {
			return duration
		}
		break
	case config.EB_Moving:
		duration += config.TRAVEL_TIME_S / 2
		testElev.Floor += int(testElev.Dirn)
		break
	case config.EB_DoorOpen:
		duration -= config.DOOR_OPEN_TIME_S / 2
	}

	for true {
		if requests.Requests_shouldStop(testElev) {
			testElev = requests.Requests_clearAtCurrentFloor(testElev, nil)
			duration += config.DOOR_OPEN_TIME_S
			a = requests.Requests_nextAction(testElev)
			testElev.Dirn = a.Dirn
			if testElev.Dirn == config.MD_Stop {
				return duration
			}
		}
		testElev.Floor += int(testElev.Dirn)
		duration += config.TRAVEL_TIME_S
	}
	return duration
}
