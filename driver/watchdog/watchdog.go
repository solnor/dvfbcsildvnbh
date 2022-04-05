package watchdog

import (
	nodeConfig "driver/config"
	elevConfig "elevator/config"
	"fmt"
	"time"
)

func Watchdog(elev *elevConfig.Elevator) {
	var timer *time.Timer
	timer = time.NewTimer(time.Duration(1 * time.Second))
	timer.Stop()
	const interval = 150
	lastFloor := elev.Floor
	lastBehaviour := elev.Behaviour
	for {
		select {
		case <-timer.C:
			fmt.Println("Node unavailable")
			nodeConfig.ThisNode.Available = false
		case <-time.After(interval * time.Millisecond):
		}
		if elev.Floor != lastFloor {
			timer.Reset(10 * time.Second)
			nodeConfig.ThisNode.Available = true
			lastFloor = elev.Floor
		} else if elev.Behaviour != lastBehaviour {
			timer.Reset(10 * time.Second)
			nodeConfig.ThisNode.Available = true
			lastBehaviour = elev.Behaviour
		} else if elev.Behaviour == elevConfig.EB_Idle {
			timer.Reset(10 * time.Second)
			nodeConfig.ThisNode.Available = true
		}
		time.Sleep(200 * time.Millisecond)
	}
}
