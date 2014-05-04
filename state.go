package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/bdon/go.gtfs"
	"github.com/bdon/go.nextbus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The instantaneous state of a vehicle.
type VehicleState struct {
	Time  int `json:"time"`
	Index int `json:"index"`

	// Used to filter for diff updates.
	TimeAdded int `json:"-"`

	// Compare strings instead of floats
	// Why? I don't know. Is float comparison reliable?
	LatString string `json:"-"`
	LonString string `json:"-"`
}

// One inbound or outbound run of a vehicle
type VehicleRun struct {
	VehicleId string            `json:"vehicle_id"`
	StartTime int               `json:"-"`
	Dir       nextbus.Direction `json:"dir"`

	States []VehicleState `json:"states"`
}

// The entire state of the system is a list of vehicle runs.
// It also has bookkeeping so it knows how to add an observation to the state.
type RouteState struct {
	// Identifier is vehicleid_timestamp, where timestamp is when run first appeared
	Runs map[string]*VehicleRun

	//Bookkeeping for vehicle ID to current run.
	CurrentRuns map[string]*VehicleRun
	Referencer  Referencer
	Id          string
}

// since maps are not threadsafe -
// a Mutex needs to be held when writing in a new nextbus response.
// This is at the Agency level.
type AgencyState struct {
	RouteStates map[string]*RouteState
	Feed        gtfs.Feed

	Mutex  sync.RWMutex
	ticker *time.Ticker
}

func NewAgencyState(feed gtfs.Feed) *AgencyState {
	retval := AgencyState{Feed: feed, Mutex: sync.RWMutex{}}
	retval.RouteStates = make(map[string]*RouteState)
	return &retval
}

func (a AgencyState) NewRouteState(routeTag string) (*RouteState, bool) {
	retval := RouteState{Id: routeTag}
	retval.Runs = map[string]*VehicleRun{}
	retval.CurrentRuns = make(map[string]*VehicleRun)

	log.Printf("looking up %s", routeTag)
	route := a.Feed.RouteByShortName(routeTag)
	longestShape := route.LongestShape()
	if longestShape == nil {
		log.Printf("Couldn't find %s", routeTag)
		return nil, false
	}

	coords := longestShape.Coords
	retval.Referencer = NewReferencer(coords)
	return &retval, true
}

func (s VehicleState) Lat() float64 {
	f, _ := strconv.ParseFloat(s.LatString, 64)
	return f
}

func (s VehicleState) Lon() float64 {
	f, _ := strconv.ParseFloat(s.LonString, 64)
	return f
}

func newToken(vehicleId string, timestamp int) string {
	time := fmt.Sprintf("%d", timestamp)
	return strings.Join([]string{vehicleId, time}, "_")
}

// Must be called in chronological order
func (a *AgencyState) AddResponse(foo nextbus.Response, unixtime int) {
	for _, report := range foo.Reports {
		routeTag := report.RouteTag
		s, ok := a.RouteStates[routeTag]
		if !ok {
			s, ok = a.NewRouteState(routeTag)
			if !ok {
				log.Printf("BAILING OUT")
				continue
			}
			a.RouteStates[routeTag] = s
		}

		if report.LeadingVehicleId != "" {
			continue
		}
		if report.DirTag == "" {
			continue
		}
		if report.LatString == "" || report.LonString == "" {
			continue
		}

		if s == nil {
			continue
		}
		index := s.Referencer.Reference(report.Lat(), report.Lon())
		newState := VehicleState{Index: index, Time: unixtime - report.SecsSinceReport,
			LatString: report.LatString, LonString: report.LonString,
			TimeAdded: unixtime}

		c, ok := s.CurrentRuns[report.VehicleId]

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

func (s *RouteState) After(time int) map[string]VehicleRun {
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

func (a *AgencyState) Start() {

	a.ticker = time.NewTicker(10 * time.Second)

	tick := func(unixtime int) {
		log.Println("Fetching from NextBus...")
		response := nextbus.Response{}
		//Accept-Encoding: gzip, deflate
		get, err := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&t=0")
		if err != nil {
			log.Println(err)
			return
		}
		defer get.Body.Close()
		str, _ := ioutil.ReadAll(get.Body)
		xml.Unmarshal(str, &response)

		a.Mutex.Lock()
		a.AddResponse(response, unixtime)
		a.Mutex.Unlock()
		log.Println("Done Fetching.")
	}

	go func() {
		for {
			select {
			case t := <-a.ticker.C:
				tick(int(t.Unix()))
			}
		}
	}()

	go tick(int(time.Now().Unix()))
}

func (a *AgencyState) Persist(p string) {
	//Mkdirp in history/year/month/day
	fmt.Println("DUMP")
	os.Mkdir(path.Join(p, "sf-muni"), 0755)
	os.Mkdir(path.Join(p, "sf-muni/2014"), 0755)
	os.Mkdir(path.Join(p, "sf-muni/2014/05"), 0755)
	os.Mkdir(path.Join(p, "sf-muni/2014/05/03"), 0755)

	a.Mutex.RLock()
	for k, s := range a.RouteStates {
		filename := fmt.Sprintf("%s.json", k)
		file, _ := os.Create(path.Join(p, "/sf-muni/2014/05/03", filename))
		result, _ := json.Marshal(s)
		file.WriteString(string(result))
	}
	a.Mutex.RUnlock()
}

func (a *AgencyState) Restore(p string) {
	fmt.Println("RESTORE")
	// glob all files and return one agency state.
	// need to create current routes
	files, _ := filepath.Glob(path.Join(p, "sf-muni/2014/05/03/*.json"))

	for _, f := range files {
		desc, _ := ioutil.ReadFile(f)
		r := RouteState{}
		json.Unmarshal(desc, &r)
		a.RouteStates[r.Id] = &r
	}
}
