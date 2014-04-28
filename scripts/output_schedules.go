package main

import (
  "encoding/csv"
  "encoding/json"
  "os"
  "fmt"
  "io"
  "strconv"
  "strings"
	"github.com/bdon/transit_timelines/linref"
)

type TripStop struct {
  Time int `json:"time"`
  Index float64 `json:"index"`
}

type Trip struct {
  TripId string `json:"trip_id"`
  Stops []TripStop `json:"stops"`
  Dir string `json:"dir"`
}

// route 1093

// we only care about these stops (OB/IB)
// Ocean Beach 7219 / 5223
// Judah/Sunset 5224 / 5225
// Judah/19th 5199 / 5200
// Carl/Hillway 3912 / 3913
// Duboce/Church 4447 / 4448
// VanNess 6996 / 5419
// Embarcadero 7217 / 6992
// King/4th  5240 / 5239

func main() {

  rawStopList := []string{"7219","5223","5224","5225","5199","5200","3912","3913",
                  "4447","4448","6996","5419","7217","6992","5240","5239"}
  stopList := map[string]bool{}
  stopIndexes := map[string]float64{}
  for _, r := range rawStopList {
    stopList[r] = true
  }
	nReferencer := linref.NewReferencer("102909")

  stopsFile, _ := os.Open("muni_gtfs/stops.txt")
	defer stopsFile.Close()
	stopsReader := csv.NewReader(stopsFile)
	stopsReader.TrailingComma = true

  for {
	  record, err := stopsReader.Read()
	  if err == io.EOF {
      break
	  }

    if _, ok := stopList[record[0]]; ok {
			stop_lat, _ := strconv.ParseFloat(record[3], 64)
			stop_lon, _ := strconv.ParseFloat(record[4], 64)
			index := nReferencer.Reference(stop_lat, stop_lon)
      stopIndexes[record[0]] = index
    }
  }

  // done initializing stops.

  trips := []Trip{}

  service := os.Args[1]
  trips = populateTrips(trips, service, "0", stopList, stopIndexes)
  trips = populateTrips(trips, service, "1", stopList, stopIndexes)

  marshalled, _ := json.Marshal(trips)
  fmt.Printf(string(marshalled))
}

func populateTrips(trips []Trip, serviceId string, directionId string, stopList map[string]bool, stopIndexes map[string]float64) []Trip {
  tripIds := map[string]bool{}
	tripsFile, _ := os.Open("muni_gtfs/trips.txt")
	defer tripsFile.Close()
	reader := csv.NewReader(tripsFile)
	reader.TrailingComma = true

	for {
		record, err := reader.Read()
    if err == io.EOF {
      break
    }
		if record[0] == "1093" && record[1] == serviceId && record[4] == directionId {
      tripIds[record[2]] = true
		}
	}

	stopTimesFile, _ := os.Open("muni_gtfs/stop_times.txt")
	defer stopTimesFile.Close()
	stopTimesReader := csv.NewReader(stopTimesFile)
	stopTimesReader.TrailingComma = true

  var currentTrip Trip
  sentinel := false

	for {
		record, err := stopTimesReader.Read()
		if err == io.EOF {
			break
		}

		if _, ok := tripIds[record[0]]; ok {
       if _, ok := stopList[record[3]]; ok {
        if currentTrip.TripId != record[0] {
          //encountered a new trip
          if sentinel {
            trips = append(trips, currentTrip)
          }
          currentTrip = Trip{TripId:record[0],Dir:directionId}
          currentTrip.Stops = []TripStop{}
          sentinel = true
        }
        currentTrip.Stops = append(currentTrip.Stops, TripStop{Time:hstoi(record[1]),Index:stopIndexes[record[3]]})
      }
    }
	}
  
  return trips
}

func hstoi(str string) int {
	components := strings.Split(str, ":")
	hour, _ := strconv.Atoi(components[0])
	min, _ := strconv.Atoi(components[1])
	sec, _ := strconv.Atoi(components[2])
	retval := hour*60*60 + min*60 + sec
	return retval
}
