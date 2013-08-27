package main

import (
  "fmt"
  "net/http"
  "github.com/bdon/jklmnt/nextbus"
  "encoding/json"
  "log"
  "time"
  "sync"
  "io/ioutil"
  "encoding/xml"
  "github.com/bdon/jklmnt/linref"
)

type VehicleState struct {
  Time int `json:"time"`
  Index float64 `json:"index"`
  LatString string `json:"-"`
  LonString string `json:"-"`
}

type SystemState struct {
  Map map[string][]VehicleState
  Mutex sync.RWMutex
}

func NewSystemState() SystemState {
  retval := SystemState{}
  retval.Map = make(map[string][]VehicleState)
  retval.Mutex = sync.RWMutex{}
  return retval
}

func (state SystemState) handler(w http.ResponseWriter, r *http.Request) {
    state.Mutex.RLock()
    result, err := json.Marshal(state.Map)
    if err != nil {
      log.Println(err)
    }
    state.Mutex.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, string(result))
}

func (state SystemState) Tick(unixtime int) {
  log.Println("Fetching from NextBus...")
  foo := nextbus.NextBusResponse{}
  xml.Unmarshal(getXML(), &foo)
  nReferencer := linref.NewReferencer("102909")

  state.Mutex.Lock()
  for _, report := range foo.Reports {
    theslice := state.Map[report.Id]
    if report.LeadingVehicleId != "" {
      continue
    }
    if len(theslice) > 0 && report.LatString == theslice[len(theslice)-1].LatString &&
       report.LonString == theslice[len(theslice)-1].LonString {
      continue
    }
    index := nReferencer.Reference(report.Lat(), report.Lon())
    state.Map[report.Id] = append(state.Map[report.Id], VehicleState{Index:index, Time:unixtime - report.SecsSinceReport,LatString:report.LatString, LonString:report.LonString})
  }
  state.Mutex.Unlock()
  log.Println("Done Fetching.")
}

func getXML() []byte {
  resp, _ := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&r=N&t=0")
  defer resp.Body.Close()
  str, _ := ioutil.ReadAll(resp.Body)
  return str
}

func main() {
  state := NewSystemState()
  ticker := time.NewTicker(10 * time.Second)
  go func() {
    for {
      select {
        case t := <-ticker.C:
          state.Tick(int(t.Unix()))
      }
    }
  }()
  go state.Tick(int(time.Now().Unix()))
  //log.Println(nextbus.ResponseFromFile("N/1377452461.xml"))
  http.HandleFunc("/", state.handler)
  log.Println("Serving on port 8080.")
  http.ListenAndServe(":8080", nil)
}
