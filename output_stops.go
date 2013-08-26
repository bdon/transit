package main

import (
  "encoding/csv"
  "os"
  "fmt"
  "io"
  "github.com/bdon/jklmnt/linref"
  "strconv"
)

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
  fmt.Printf("trip id: %s\n", tripId)

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
  fmt.Printf("stop ids: %s\n", stopMap)

  // create a linear referencer
  nReferencer := linref.NewReferencer("102909")

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
     fmt.Println(record[1])
     stop_lat, _ := strconv.ParseFloat(record[3],64)
     stop_lon, _ := strconv.ParseFloat(record[4],64)
     fmt.Println(nReferencer.Reference(stop_lat, stop_lon))
    }
  }
}
