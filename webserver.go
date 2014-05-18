package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func Webserver(agencyState *AgencyState, allowAll bool) {

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
    log.Println("READ LOCK")

		var result []byte
    var runs interface{} //hack hack
    var ok bool
		agencyState.Mutex.RLock()
		if time > 0 {
      runs, ok = agencyState.RunsAfter(route, time)
    } else {
      runs, ok = agencyState.Runs(route)
    }
		agencyState.Mutex.RUnlock()

    if ok {
      result, _ = json.Marshal(runs)
    } else {
      http.Error(w, "BAD REQUEST", http.StatusBadGateway)
      return
    }

    log.Println("READ UNLOCKED")
		w.Header().Set("Content-Type", "application/json")
		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		fmt.Fprintf(w, string(result))
	}

	http.HandleFunc("/locations.json", handler)
	http.HandleFunc("/", healthHandler)
	log.Println("Serving on port 8080.")
	http.ListenAndServe(":8080", nil)
}
