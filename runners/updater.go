package runners

import (
	"fmt"
	"sync"

	nmea "github.com/Anbcorp/sa_nmea"
)

type Updater struct {
	Task
	Boat  *nmea.NMEABoat
	mutex *sync.RWMutex
}

func NewUpdater(d int, b *nmea.NMEABoat, m *sync.RWMutex) Updater {
	var nu Updater
	nu.Task = NewTask(d)
	nu.Boat = b
	nu.mutex = m

	return nu
}

func (upd Updater) Update() {
	upd.mutex.Lock()
	newHdg := int(upd.Boat.Hdg+45) % 360
	newCog := int(upd.Boat.Cog+45) % 360
	upd.Boat.Hdg = float64(newHdg)
	upd.Boat.Cog = float64(newCog)
	fmt.Println(upd.Boat.Cog)
	upd.mutex.Unlock()
}

func (upd Updater) Run() {
	for {
		select {
		case <-upd.Done:
			upd.Ticker.Stop()
			return
		case <-upd.Ticker.C:
			upd.Update()
		}
	}
}

func (upd Updater) Stop() {
	upd.Done <- true
}
