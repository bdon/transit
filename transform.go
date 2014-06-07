package main

import (
	"encoding/json"
	"fmt"
	"github.com/bdon/go.gtfs"
	"log"
	"os"
	"path"
)

type StopRepr struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

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
	Stops []StopRepr `json:"stops"`
	Trips []TripRepr `json:"trips"`
}

type RouteRepr struct {
	Id        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
}

func EmitRoot(feed gtfs.Feed) {
	output := []RouteRepr{}
	for _, route := range feed.Routes {
		r := RouteRepr{Id: route.Id, ShortName: route.ShortName, LongName: route.LongName}
		output = append(output, r)
	}
	marshalled, _ := json.MarshalIndent(output, "", "  ")
	fmt.Printf(string(marshalled))
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

func EmitSchedules(feed gtfs.Feed) {
	perRoute(feed, "schedules", func(route *gtfs.Route) (string, bool) {
		// TODO: handle missing shape
		shape := route.LongestShape()
		if shape == nil {
			return "", false
		}
		referencer := NewReferencer(shape.Coords)

		trips := []TripRepr{}
		for _, trip := range route.Trips {
			// TODO: we're only caring about weekdays for now
			if trip.Service != "1" {
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

		stops := []StopRepr{}
		for _, stop := range route.Stops() {
			index := referencer.Reference(stop.Coord.Lat, stop.Coord.Lon)
			stops = append(stops, StopRepr{Index: index, Name: stop.Name})
		}

		schedule := ScheduleRepr{Trips: trips, Stops: stops}

		marshalled, _ := json.Marshal(schedule)
		return string(marshalled), true
	})
}
