package runners

import (
	"sync"

	nmea "github.com/Anbcorp/sa_nmea"
	"github.com/tidwall/geodesic"
)

type Deadrecker struct {
	Task
	Boat  *nmea.NMEABoat
	mutex *sync.RWMutex
}

func NewDeadrecker(d int, b *nmea.NMEABoat, m *sync.RWMutex) Deadrecker {
	var nu Deadrecker
	nu.Task = NewTask(d)
	nu.Boat = b
	nu.mutex = m

	return nu
}

func (dedrek Deadrecker) Update() {
	var newLat, newLon float64
	dedrek.mutex.Lock()
	geodesic.WGS84.Direct(
		dedrek.Boat.Latitude, dedrek.Boat.Longitude,
		dedrek.Boat.Cog, dedrek.Boat.Sog/360,
		&newLat, &newLon, nil)
	dedrek.Boat.Latitude = newLat
	dedrek.Boat.Longitude = newLon
	dedrek.mutex.Unlock()
}

func (dedrek Deadrecker) Run() {
	for {
		select {
		case <-dedrek.Done:
			dedrek.Ticker.Stop()
			return
		case <-dedrek.Ticker.C:
			dedrek.Update()
		}
	}
}

func (dedrek Deadrecker) Stop() {
	dedrek.Done <- true
}
