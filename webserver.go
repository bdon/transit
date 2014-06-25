package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func Webserver(agencyState *AgencyState) {
	result, _ := json.Marshal(DateRangeFs("static/history"))

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
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

		resolve, rok := agencyState.Names.GtoN(route)
		if rok {
			route = resolve
		}

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

		w.Header().Set("Content-Type", "application/json")
		// You should override this in nginx
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, string(result))
	}

	http.HandleFunc("/locations.json", handler)
	http.HandleFunc("/", healthHandler)
	http.ListenAndServe(":8080", nil)
}
