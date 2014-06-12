package main

import (
	"encoding/json"
	"fmt"
	"github.com/bdon/go.gtfs"
	"log"
	"os"
	"path"
	"sort"
)

type StopRepr struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

type StopByIndex []StopRepr

func (a StopByIndex) Len() int           { return len(a) }
func (a StopByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StopByIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

type RouteByShortName []RouteRepr

func (a RouteByShortName) Len() int           { return len(a) }
func (a RouteByShortName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RouteByShortName) Less(i, j int) bool { return a[i].ShortName < a[j].ShortName }

type StopTimeRepr struct {
	Time  int `json:"time"`
	Index int `json:"index"`
}

type TripRepr struct {
	TripId    string         `json:"trip_id"`
	StopTimes []StopTimeRepr `json:"stops"`
	Dir       string         `json:"dir"`
}

type ScheduleRepr struct {
	Stops     []StopRepr `json:"stops"`
	Trips     []TripRepr `json:"trips"`
	Headsigns []string   `json:"headsigns"`
}

type Root struct {
	Calendar []string    `json:"calendar"`
	Routes   []RouteRepr `json:"routes"`
}

type RouteRepr struct {
	Id        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
}

// emit the calendar, and GTFS routes
func EmitRoot(feed gtfs.Feed) {
	output := []RouteRepr{}
	for _, route := range feed.Routes {
		r := RouteRepr{Id: route.Id, ShortName: route.ShortName, LongName: route.LongName}
		output = append(output, r)
	}

	fmt.Println("Writing ", "static/routes.json")
	file, err := os.Create("static/routes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	sort.Sort(RouteByShortName(output))

	root := Root{Routes: output, Calendar: feed.Calendar()}
	marshalled, _ := json.MarshalIndent(root, "", "  ")
	file.WriteString(string(marshalled))
}

func perRoute(feed gtfs.Feed, dirname string, f func(*gtfs.Route) (string, bool)) {
	for _, route := range feed.Routes {
		foo := fmt.Sprintf("%s.json", route.Id)
		_ = os.Mkdir(path.Join("static", dirname), 0755)
		filename := path.Join("static", dirname, foo)
		fmt.Println("Writing ", filename)
		file, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		str, ok := f(route)
		if ok {
			file.WriteString(str)
		}
	}
}

func EmitStops(feed gtfs.Feed) {
	perRoute(feed, "stops", func(route *gtfs.Route) (string, bool) {
		// TODO: handle missing shape
		shape := route.LongestShape()
		if shape == nil {
			return "", false
		}
		referencer := NewReferencer(shape.Coords)

		stops := []StopRepr{}
		for _, stop := range route.Stops() {
			index := referencer.Reference(stop.Coord.Lat, stop.Coord.Lon)
			stops = append(stops, StopRepr{Index: index, Name: stop.Name})
		}

		sort.Sort(StopByIndex(stops))

		schedule := ScheduleRepr{Stops: stops, Headsigns: route.Headsigns()}
		marshalled, _ := json.Marshal(schedule)
		return string(marshalled), true
	})
}

func jsonForRouteAndServiceId(route *gtfs.Route, serviceId string) (string, bool) {
	// TODO: handle missing shape
	shape := route.LongestShape()
	if shape == nil {
		return "", false
	}
	referencer := NewReferencer(shape.Coords)

	trips := []TripRepr{}
	for _, trip := range route.Trips {
		if trip.Service != serviceId {
			continue
		}
		tripRepr := TripRepr{TripId: trip.Id, Dir: trip.Direction}
		for _, stoptime := range trip.StopTimes {
			index := referencer.Reference(stoptime.Stop.Coord.Lat, stoptime.Stop.Coord.Lon)
			newStopTime := StopTimeRepr{Time: stoptime.Time, Index: index}
			tripRepr.StopTimes = append(tripRepr.StopTimes, newStopTime)
		}
		trips = append(trips, tripRepr)
	}

	marshalled, _ := json.Marshal(trips)
	return string(marshalled), true
}

func EmitSchedules(feed gtfs.Feed) {
	os.Mkdir("static/schedules", 0755)
	os.Mkdir("static/schedules/1", 0755)
	os.Mkdir("static/schedules/2", 0755)
	os.Mkdir("static/schedules/3", 0755)
	perRoute(feed, "schedules/1", func(route *gtfs.Route) (string, bool) {
		return jsonForRouteAndServiceId(route, "1")
	})
	perRoute(feed, "schedules/2", func(route *gtfs.Route) (string, bool) {
		return jsonForRouteAndServiceId(route, "2")
	})
	perRoute(feed, "schedules/3", func(route *gtfs.Route) (string, bool) {
		return jsonForRouteAndServiceId(route, "3")
	})
}
