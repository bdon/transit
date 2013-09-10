package state

import (
	"fmt"
	"github.com/bdon/jklmnt/linref"
	"github.com/bdon/jklmnt/nextbus"
	"strconv"
	"strings"
)

// The instantaneous state of a vehicle as returned by NextBus
type VehicleState struct {
	Time      int     `json:"time"`
	Index     float64 `json:"index"`
	TimeAdded int     `json:"-"`

	LatString string `json:"-"`
	LonString string `json:"-"`
}

// One inbound or outbound run of a vehicle
type VehicleRun struct {
	VehicleId string            `json:"vehicle_id"`
	StartTime int               `json:"-"`
	Dir       nextbus.Direction `json:"dir"`
	States    []VehicleState    `json:"states"`
}

// The entire state of the system is a list of vehicle runs.
// It also has bookkeeping so it knows how to add an observation to the state.
// And synchronization primitives.
type SystemState struct {
	// has an identifier which is vehicleid+timestamp
	Runs map[string]*VehicleRun

	//Bookkeeping for vehicle ID to current run.
	CurrentRuns map[string]*VehicleRun
	Referencer  linref.Referencer
}

func NewSystemState() *SystemState {
	retval := SystemState{}
	retval.Runs = map[string]*VehicleRun{}
	retval.CurrentRuns = make(map[string]*VehicleRun)
	retval.Referencer = linref.NewReferencer("102909")
	return &retval
}

func (s VehicleState) Lat() float64 {
	f, _ := strconv.ParseFloat(s.LatString, 64)
	return f
}

func (s VehicleState) Lon() float64 {
	f, _ := strconv.ParseFloat(s.LonString, 64)
	return f
}

// simplify needs to act on arbitrary objects
func (r *VehicleRun) Simplify() {
}

func newToken(vehicleId string, timestamp int) string {
	time := fmt.Sprintf("%d", timestamp)
	return strings.Join([]string{vehicleId, time}, "_")
}

// Must be called in chronological order
func (s *SystemState) AddResponse(foo nextbus.Response, unixtime int) {
	for _, report := range foo.Reports {
		if report.LeadingVehicleId != "" {
			continue
		}
		if report.DirTag == "" {
			continue
		}

		index := s.Referencer.Reference(report.Lat(), report.Lon())
		// cull data on first and last stops
		//if index > 0.9975 || index < 0.0268 {
		//  continue
		//}
		newState := VehicleState{Index: index, Time: unixtime - report.SecsSinceReport,
			LatString: report.LatString, LonString: report.LonString,
			TimeAdded: unixtime}

		c := s.CurrentRuns[report.VehicleId]
		if c != nil {
			lastState := c.States[len(c.States)-1]

			if newState.Time-lastState.Time > 900 || report.Dir() != c.Dir {
				// create a new Run
				startTime := unixtime - report.SecsSinceReport
				newRun := VehicleRun{VehicleId: report.VehicleId, Dir: report.Dir(), StartTime: startTime}
				newRun.States = append(newRun.States, newState)
				s.Runs[newToken(report.VehicleId, startTime)] = &newRun
				s.CurrentRuns[newRun.VehicleId] = &newRun

			} else if lastState.LatString != newState.LatString || lastState.LonString != newState.LonString {
				c.States = append(c.States, newState)
			}
		} else {
			startTime := unixtime - report.SecsSinceReport
			newRun := VehicleRun{VehicleId: report.VehicleId, Dir: report.Dir(), StartTime: startTime}
			newRun.States = append(newRun.States, newState)
			s.Runs[newToken(report.VehicleId, startTime)] = &newRun
			s.CurrentRuns[newRun.VehicleId] = &newRun
		}
	}
}

func (s *SystemState) After(time int) map[string]VehicleRun {
	filtered := map[string]VehicleRun{}

	for token, run := range s.Runs {
		for _, s := range run.States {
			if s.TimeAdded >= time {
				if _, ok := filtered[token]; ok {
					foo := filtered[token]
					foo.States = append(filtered[token].States, s)
					filtered[token] = foo
				} else {
					foo := *run
					foo.States = []VehicleState{s}
					filtered[token] = foo
				}
			}
		}
	}
	return filtered
}

func (s *SystemState) DeleteOlderThan(duration int, currentTime int) {
	// replace runs with a filtered list
	replacementList := map[string]*VehicleRun{}
	for key, run := range s.Runs {
		if run.StartTime > currentTime-duration {
			replacementList[key] = run
		}
	}

	s.Runs = replacementList

	replacementCurrent := map[string]*VehicleRun{}
	for key, run := range s.CurrentRuns {
		if run.StartTime > currentTime-duration {
			replacementCurrent[key] = run
		}
	}

	s.CurrentRuns = replacementCurrent
}
