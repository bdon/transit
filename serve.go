package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/bdon/jklmnt/nextbus"
	"github.com/bdon/jklmnt/state"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	s := state.NewSystemState()
	ticker := time.NewTicker(10 * time.Second)
	mutex := sync.RWMutex{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		result, err := json.Marshal(s.Runs)
		if err != nil {
			log.Println(err)
		}
		mutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	}

	tick := func(unixtime int) {
		log.Println("Fetching from NextBus...")
		response := nextbus.Response{}
		get, _ := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&r=N&t=0")
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

	// do the initial thing
	go tick(int(time.Now().Unix()))

	http.HandleFunc("/", handler)
	log.Println("Serving on port 8080.")
	http.ListenAndServe(":8080", nil)
}
