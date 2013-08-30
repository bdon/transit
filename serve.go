package main

import (
  "net/http"
  "log"
  "time"
  "github.com/bdon/jklmnt/state"
  "github.com/bdon/jklmnt/nextbus"
  "encoding/xml"
  "io/ioutil"
)

func Tick(s *state.SystemState, unixtime int) {
  log.Println("Fetching from NextBus...")
  response := nextbus.Response{}
  get, _ := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&r=N&t=0")
  defer get.Body.Close()
  str, _ := ioutil.ReadAll(get.Body)
  xml.Unmarshal(str, &response)

  s.Mutex.Lock()
  s.AddResponse(response, unixtime)
  log.Println(len(s.Runs))
  s.Mutex.Unlock()
  log.Println("Done Fetching.")
}

func main() {
  s := state.NewSystemState()
  ticker := time.NewTicker(10 * time.Second)
  go func() {
    for {
      select {
        case t := <-ticker.C:
          Tick(s, int(t.Unix()))
      }
    }
  }()

  // do the initial thing
  go Tick(s, int(time.Now().Unix()))

  http.HandleFunc("/", s.Handler)
  log.Println("Serving on port 8080.")
  http.ListenAndServe(":8080", nil)
}
