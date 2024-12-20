package runners

import (
	"fmt"
	"sync"
	"time"

	nmea "github.com/Anbcorp/sa_nmea"
)

type Viewer struct {
	Task
	Boat  *nmea.NMEABoat
	mutex *sync.RWMutex
}

func NewViewer(d int, b *nmea.NMEABoat, m *sync.RWMutex) Viewer {
	var nu Viewer
	nu.Task = NewTask(d)
	nu.Boat = b
	nu.mutex = m

	return nu
}

func (view Viewer) Update() {
	view.mutex.RLock()
	fmt.Println(time.Now(), view.Boat.Latitude, view.Boat.Longitude, view.Boat.Cog)
	view.mutex.RUnlock()
}

func (view Viewer) Run() {
	for {
		select {
		case <-view.Done:
			view.Ticker.Stop()
			return
		case <-view.Ticker.C:
			view.Update()
		}
	}
}

func (view Viewer) Stop() {
	view.Done <- true
}
