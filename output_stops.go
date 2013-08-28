package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/bdon/jklmnt/linref"
	"io"
	"os"
	"strconv"
	"strings"
)

type StopRepr struct {
	Index float64 `json:"index"`
	Name  string  `json:"name"`
	MyName  string  `json:"my_name"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}

func main() {
	// Find the first trip for a shape
	tripsFile, _ := os.Open("muni_gtfs/trips.txt")
	defer tripsFile.Close()
	reader := csv.NewReader(tripsFile)
	reader.TrailingComma = true
	var tripId string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			fmt.Println("No Record Found")
			break
		}
		if record[6] == "102909" {
			tripId = record[2]
			break
		}
	}
	//fmt.Printf("trip id: %s\n", tripId)

	// Create a map of stop ids for that trip
	stopTimesFile, _ := os.Open("muni_gtfs/stop_times.txt")
	defer stopTimesFile.Close()
	stopTimesReader := csv.NewReader(stopTimesFile)
	stopTimesReader.TrailingComma = true
	stopMap := make(map[string]bool)
	for {
		record, err := stopTimesReader.Read()
		if err == io.EOF {
			break
		}
		if record[0] == tripId {
			stopMap[record[3]] = true
		}
	}
	//fmt.Printf("stop ids: %s\n", stopMap)

	// create a linear referencer
	nReferencer := linref.NewReferencer("102909")

	// create an output data structure
	output := []StopRepr{}

  myNames := make(map[string]string)
  myNames["King St & 4th St"] = "King & 4th"
  myNames["King St & 2th St"] = "King & 2nd"
  myNames["The Embarcadero & Brannan St"] = "Brannan"
  //myNames["The Embarcadero & Harrison St"] = "Harrison"
  myNames["The Embarcadero & Folsom St"] = "Folsom"
  myNames["Metro Embarcadero Station"] = "Embarcadero"
  myNames["Metro Montgomery Station/Outbound"] = "Montgomery"
  myNames["Metro Powell Station/Outbound"] = "Powell"
  myNames["Metro Civic Center Station/Outbd"] = "Civic Center"
  myNames["Van Ness Station Outbound"] = "Van Ness"
  myNames["Duboce Ave & Church St"] = "Duboce & Church"
  myNames["Duboce St/Noe St/Duboce Park"] = "Duboce & Noe"
  myNames["Carl St & Cole St"] = "Carl & Cole"
  myNames["Carl St & Stanyan St"] = "Carl & Stanyan"
  //myNames["Carl St & Hillway Ave"] = "Carl & Hillway"
  myNames["Irving St & 2nd Ave"] = "Irving & 2nd"
  //myNames["Irving St & 4th Ave"] = "Irving & 4th"
  //myNames["Irving St & 7th Ave"] = "Irving & 7th"
  //myNames["Irving St & 9th Ave"] = "Irving & 9th"
  myNames["Judah St & 9th Ave"] = "Judah & 9th"
  //myNames["Judah St & Funston Ave"] = "Judah & Funston"
  //myNames["Judah St & 16th Ave"] = "Judah & 16th"
  myNames["Judah St & 19th Ave"] = "Judah & 19th"
  //myNames["Judah St & 23rd Ave"] = "Judah & 23rd"
  //myNames["Judah St & 25th Ave"] = "Judah & 25th"
  //myNames["Judah St & 28th Ave"] = "Judah & 28th"
  //myNames["Judah St & 31st Ave"] = "Judah & 31st"
  //myNames["Judah St & 34th Ave"] = "Judah & 34th"
  myNames["Judah St & Sunset Blvd"] = "Judah & Sunset"
  //myNames["Judah St & 40th Ave"] = "Judah & 40th"
  //myNames["Judah St & 43rd Ave"] = "Judah & 43rd"
  //myNames["Judah St & 46th Ave"] = "Judah & 46th"
  myNames["Judah/La Playa/Ocean Beach"] = "Ocean Beach"


	// print all stops given a list of stop IDs
	stopsFile, _ := os.Open("muni_gtfs/stops.txt")
	defer stopsFile.Close()
	stopsReader := csv.NewReader(stopsFile)
	stopsReader.TrailingComma = true
	for {
		record, err := stopsReader.Read()
		if err == io.EOF {
			break
		}
		if stopMap[record[0]] {
			newStop := StopRepr{}
			stop_lat, _ := strconv.ParseFloat(record[3], 64)
			stop_lon, _ := strconv.ParseFloat(record[4], 64)
			index := nReferencer.Reference(stop_lat, stop_lon)
			newStop.Lat = stop_lat
			newStop.Lon = stop_lon
			newStop.Index = index
			newStop.Name = strings.TrimSpace(record[1])
      newStop.MyName = myNames[strings.TrimSpace(record[1])]
			output = append(output, newStop)
		}
	}

	marshalled, _ := json.Marshal(output)
	fmt.Printf(string(marshalled))
}
