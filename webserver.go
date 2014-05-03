package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func Webserver(agencyState *AgencyState) {

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

		agencyState.Mutex.RLock()

		var result []byte
		if time > 0 {
			result, _ = json.Marshal(agencyState.RouteStates[route].After(time))
		} else {
			result, _ = json.Marshal(agencyState.RouteStates[route].Runs)
		}
		agencyState.Mutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, string(result))
	}

	http.HandleFunc("/locations.json", handler)
	http.HandleFunc("/", healthHandler)
	log.Println("Serving on port 8080.")
	http.ListenAndServe(":8080", nil)
}
