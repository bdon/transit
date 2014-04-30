package main

import (
	"encoding/json"
	"fmt"
  "github.com/bdon/transit_timelines"
  "github.com/bdon/go.gtfs"
)

type StopRepr struct {
	Index     float64 `json:"index"`
	Name      string  `json:"name"`
}

func main() {
  feed := gtfs.Load("muni_gtfs")
  route := feed.RouteByShortName("N")
	referencer := transit_timelines.NewReferencer(route.LongestShape().Coords)
	output := []StopRepr{}

  for _, stop := range route.Stops() {
	  index := referencer.Reference(stop.Coord.Lat, stop.Coord.Lon)
    output = append(output, StopRepr{Index:index,Name:stop.Name})
  }

	marshalled, _ := json.Marshal(output)
	fmt.Printf(string(marshalled))
}
