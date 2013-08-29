package state

import (
  "sync"
  "encoding/json"
  "encoding/xml"
  "github.com/bdon/jklmnt/linref"
  "github.com/bdon/jklmnt/nextbus"
  "net/http"
  "log"
  "fmt"
  "io/ioutil"
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
  Referencer linref.Referencer
}

func NewSystemState() SystemState {
  retval := SystemState{}

  // we're gonna need more data structures.
  //the underlying data structure is a list []
  // we hold a map of vehicle ids
  // with pointers into elements

  retval.Map = make(map[string][]VehicleState)
  retval.Mutex = sync.RWMutex{}
  retval.Referencer = linref.NewReferencer("102909")
  return retval
}

func (s SystemState) Handler(w http.ResponseWriter, r *http.Request) {
    s.Mutex.RLock()
    result, err := json.Marshal(s.Map)
    if err != nil {
      log.Println(err)
    }
    s.Mutex.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, string(result))
}

func (s SystemState) AddResponse(foo nextbus.NextBusResponse, unixtime int) {
  // here's the magic
  // maintain a list of current vehicle runs

  // we care about 'runs' where the direction tag is the same
  for _, report := range foo.Reports {
    theslice := s.Map[report.Id]
    if report.LeadingVehicleId != "" {
      continue
    }
    if len(theslice) > 0 && report.LatString == theslice[len(theslice)-1].LatString &&
       report.LonString == theslice[len(theslice)-1].LonString {
      continue
    }
    index := s.Referencer.Reference(report.Lat(), report.Lon())
    s.Map[report.Id] = append(s.Map[report.Id], VehicleState{Index:index, Time:unixtime - report.SecsSinceReport,LatString:report.LatString, LonString:report.LonString})
  }
}

func (s SystemState) Tick(unixtime int) {
  log.Println("Fetching from NextBus...")
  foo := nextbus.NextBusResponse{}
  resp, _ := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&r=N&t=0")
  defer resp.Body.Close()
  str, _ := ioutil.ReadAll(resp.Body)
  xml.Unmarshal(str, &foo)

  s.Mutex.Lock()
  s.AddResponse(foo, unixtime)
  s.Mutex.Unlock()
  log.Println("Done Fetching.")
}
