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
	"strconv"
	"sync"
	"time"
)

func Webserver() {
	feed := gtfs.Load("muni_gtfs", false)
	agencyState := NewAgencyState(feed)
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
		var route string

		if _, ok := r.Form["after"]; ok {
			time, _ = strconv.Atoi(r.Form["after"][0])
		}

		if _, ok := r.Form["route"]; ok {
			route = r.Form["route"][0]
		}

		mutex.RLock()

		var result []byte
		if time > 0 {
			result, _ = json.Marshal(agencyState.RouteStates[route].After(time))
		} else {
			result, _ = json.Marshal(agencyState.RouteStates[route].Runs)
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
		agencyState.AddResponse(response, unixtime)
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
				agencyState.DeleteOlderThan(60*60*12, int(t.Unix()))
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
