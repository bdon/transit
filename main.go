package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/bdon/go.gtfs"
	"github.com/bdon/go.nextbus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

//  -emitFiles DIR
//  outputs stops/schedules based on all GTFS feeds into directory DIR
//  serve these compiled files through NGINX.
//  -port PORT
// -gtfs DIR
// reads GTFS data from DIR

var emitFiles bool

func init() {
	flag.BoolVar(&emitFiles, "emitFiles", false, "emit files")
}

func main() {
	flag.Parse()
	if emitFiles {
		feed := gtfs.Load("muni_gtfs", true)
		emitStops(feed)
		emitSchedules(feed)
	} else {
		webserver()
	}
}

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

func emitStops(feed gtfs.Feed) {
	perRoute(feed, "stops", func(route *gtfs.Route) (string, bool) {
		// TODO: handle missing shape
		shape := route.LongestShape()
		if shape == nil {
			return "", false
		}
		referencer := NewReferencer(shape.Coords)

		output := []StopRepr{}
		for _, stop := range route.Stops() {
			index := referencer.Reference(stop.Coord.Lat, stop.Coord.Lon)
			output = append(output, StopRepr{Index: index, Name: stop.Name})
		}
		marshalled, _ := json.Marshal(output)
		return string(marshalled), true
	})
}

func emitSchedules(feed gtfs.Feed) {
	perRoute(feed, "schedules", func(route *gtfs.Route) (string, bool) {
		// TODO: handle missing shape
		shape := route.LongestShape()
		if shape == nil {
			return "", false
		}
		referencer := NewReferencer(shape.Coords)

		output := []TripRepr{}
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
			output = append(output, tripRepr)
		}
		marshalled, _ := json.Marshal(output)
		return string(marshalled), true
	})
}

func webserver() {
	s := NewRouteState()
	ticker := time.NewTicker(10 * time.Second)
	cleanupTicker := time.NewTicker(60 * time.Second)
	mutex := sync.RWMutex{}

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "Hello there.")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		var time int

		if _, ok := r.Form["after"]; ok {
			time, _ = strconv.Atoi(r.Form["after"][0])
		}

		mutex.RLock()

		var result []byte
		if time > 0 {
			result, _ = json.Marshal(s.After(time))
		} else {
			result, _ = json.Marshal(s.Runs)
		}
		mutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, string(result))
	}

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

		mutex.Lock()
		s.AddResponse(response, unixtime)
		mutex.Unlock()
		log.Println("Done Fetching.")
	}

	go func() {
		for {
			select {
			case t := <-ticker.C:
				tick(int(t.Unix()))
			}
		}
	}()

	go func() {
		for {
			select {
			case t := <-cleanupTicker.C:
				log.Println("Deleting runs older than 12 hours.")
				mutex.Lock()
				s.DeleteOlderThan(60*60*12, int(t.Unix()))
				mutex.Unlock()
				log.Println("Done cleaning up.")
			}
		}
	}()

	// do the initial thing
	go tick(int(time.Now().Unix()))

	http.HandleFunc("/locations.json", handler)
	http.HandleFunc("/", healthHandler)
	log.Println("Serving on port 8080.")
	http.ListenAndServe(":8080", nil)
}
