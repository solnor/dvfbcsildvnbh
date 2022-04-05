package timer

import (
	"time"
)

// var Timer = make(chan bool)

// var DoorTimer *time.Timer
//Initialize timer
var doorTimer *time.Timer

func Observer(timerStop chan bool) {
	doorTimer = time.NewTimer(time.Duration(1 * time.Second))
	doorTimer.Stop()
	for {
		<-doorTimer.C
		timerStop <- true
	}
}

// func StartTimer(seconds int64) {
// 	//Initialize timers
// 	DoorTimer = time.NewTimer(time.Duration(1000 * time.Second))

// }

func Reset(duration_s time.Duration) {
	doorTimer.Reset(duration_s)
}
