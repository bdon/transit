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
	CurrentRuns map[string]*VehicleRun `json:"-"`
	Referencer  Referencer             `json:"-"`
	Id          string                 `json:"id"`
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

func (a *AgencyState) Runs(routeTag string) (map[string]*VehicleRun, bool) {
	s, ok := a.RouteStates[routeTag]
	if ok {
		return s.Runs, true
	}
	return make(map[string]*VehicleRun), false
}

func (a *AgencyState) RunsAfter(routeTag string, unixtime int) (map[string]VehicleRun, bool) {
	s, ok := a.RouteStates[routeTag]
	if ok {
		return s.After(unixtime), true
	}
	return make(map[string]VehicleRun), false
}

func (a *AgencyState) Start() {

	a.ticker = time.NewTicker(10 * time.Second)

	tick := func(unixtime int) {
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
	log.Println("DUMP")
	t := time.Now()
	MkdirpForTime(p, t)

	a.Mutex.RLock()
	for k, s := range a.RouteStates {
		filename := fmt.Sprintf("%s.json", k)
		fullpath := filepath.Join(FilepathForTime(p, t), filename)
		file, err := os.Create(fullpath)
		if err != nil {
			log.Printf("Error: %s", err)
		}
		defer file.Close()
		result, _ := json.Marshal(s)
		_, err = file.WriteString(string(result))
		if err != nil {
			log.Printf("Error: %s", err)
		}
	}
	a.Mutex.RUnlock()
}

// TODO should probably have a Mutex here.
func (a *AgencyState) Restore(p string) {
	// glob all files and return one agency state.
	// need to create current routes

	t := time.Now()
	fp := FilepathForTime(p, t)
	files, _ := filepath.Glob(filepath.Join(fp, "/*.json"))

	for _, f := range files {
		desc, _ := ioutil.ReadFile(f)
		r := RouteState{}
		json.Unmarshal(desc, &r)

		route := a.Feed.RouteByShortName(r.Id)
		longestShape := route.LongestShape()
		if longestShape == nil {
			log.Printf("Couldn't find %s", r.Id)
			continue
		}
		coords := longestShape.Coords
		r.Referencer = NewReferencer(coords)
		a.RouteStates[r.Id] = &r

		r.CurrentRuns = make(map[string]*VehicleRun)
		// stitch together the current runs
		for _, x := range r.Runs {
			if r.CurrentRuns[x.VehicleId] == nil ||
				r.CurrentRuns[x.VehicleId].StartTime < x.StartTime {
				r.CurrentRuns[x.VehicleId] = x
			}
		}
	}
}

func FilepathForTime(p string, t time.Time) string {
	y := strconv.FormatInt(int64(t.Year()), 10)
	m := strconv.FormatInt(int64(t.Month()), 10)
	d := strconv.FormatInt(int64(t.Day()), 10)
	return filepath.Join(p, "/", y, "/", m, "/", d)
}

func MkdirpForTime(p string, t time.Time) {
	y := strconv.FormatInt(int64(t.Year()), 10)
	m := strconv.FormatInt(int64(t.Month()), 10)
	d := strconv.FormatInt(int64(t.Day()), 10)

	os.Mkdir(filepath.Join(p, "/", y), 0755)
	os.Mkdir(filepath.Join(p, "/", y, "/", m), 0755)
	os.Mkdir(filepath.Join(p, "/", y, "/", m, "/", d), 0755)
}
