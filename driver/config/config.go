package config

import (
	elevConfig "elevator/config"
	"elevator/fsm"
	"sync"
	"time"
)

type Node struct {
	Id        string
	Available bool
	Elevator  elevConfig.Elevator
}

// var KnownNodes = make([]*Node, 0)
var KnownNodesMutex = sync.RWMutex{}
var KnownNodesTable map[string]*Node

var ThisNode Node

func Node_Init(id string) {
	KnownNodesTable = make(map[string]*Node)
	ThisNode = NewNode(id)
	ThisNode.Elevator = fsm.Elevator1
	// nodeConfig.KnownNodes = make(map[string])
	// nodeConfig.KnownNodes = append(nodeConfig.KnownNodes, &thisNode) //MOVE THIS
	KnownNodesMutex.Lock()
	KnownNodesTable[id] = &ThisNode
	KnownNodesMutex.Unlock()
}

func NewNode(id string) Node {
	var n Node
	n.Id = id
	n.Available = true
	n.Elevator = elevConfig.NewElevator()
	return n
}

type Order struct {
	// MessageFrom string
	SenderId   string
	AssignedId string
	Request    elevConfig.ButtonEvent
	Timestamp  time.Time
	Cost       int64
	Acks       []string
	// OneAckIsEnough bool
	State OrderType
}

type OrderEvent struct {
	Request   elevConfig.ButtonEvent
	Confirmed bool
}

type OrderType int

const (
	Order_Cleared   = -1
	Order_New       = 0
	Order_Ack       = 1
	Order_Confirmed = 2
)

// type[new, ack, confirmed, cleared]

type Acks struct {
	NodeId string
	AckId  int
}

// func FindNodeWithId(id string) Node {
// 	for _,node := range(KnownNodes) {

// 	}
// }