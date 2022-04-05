package config

var NumFloors int = 4
var N_FLOORS int = 4
var N_BUTTONS int = 3
var DOOR_OPEN_TIME_S int64 = 3
var TRAVEL_TIME_S int64 = 5

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// func (btn ButtonEvent) int() int {
// 	return int(c)
// }

func ToButtonType(btn int) ButtonType {
	switch btn {
	case 0:
		return BT_HallUp
	case 1:
		return BT_HallDown
	}
	return BT_Cab
}

type ElevatorBehaviour int

const (
	EB_Idle     = 0
	EB_DoorOpen = 1
	EB_Moving   = 2
)

type ClearRequestVariant int

const (
	CV_All    = 0
	CV_InDirn = 1
)

type Elevator struct {
	Floor       int
	Dirn        MotorDirection
	Requests    [][3]int
	Behaviour   ElevatorBehaviour
	Obstruction bool

	Config Config
}

type Action struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

func NewElevator() Elevator {
	var e Elevator
	e.Requests = make([][3]int, NumFloors)
	return e
}

func DupElevator(e Elevator) Elevator {
	e2 := NewElevator()
	e2.Floor = e.Floor
	e2.Dirn = e.Dirn
	e2.Behaviour = e.Behaviour
	e2.Obstruction = e.Obstruction
	e2.Config = e.Config
	for k, v := range e.Requests {
		e2.Requests[k] = v
	}
	return e2
}

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  float64
}
type ElevOutputDevice struct {
}
