package main

import (
  "encoding/xml"
  "os"
  "io/ioutil"
  "log"
  "github.com/bdon/jklmnt/linref"
)

type NextBusVehicleReport struct {
  Id string `xml:"id,attr"`
  DirTag string `xml:"dirTag,attr"`
  Lat float64 `xml:"lat,attr"`
  Lon float64 `xml:"lon,attr"`
  SecsSinceReport int `xml:"secsSinceReport,attr"`
}

type NextBusResponse struct {
  Reports []NextBusVehicleReport `xml:"vehicle"`
}

func main() {
  file, err := os.Open("N/1377452461.xml")
  if err != nil {
    log.Fatal(err)
  }
  foo := NextBusResponse{}
  body, err := ioutil.ReadAll(file)
  if err != nil {
    log.Fatal(err)
  }
  xml.Unmarshal(body, &foo)
  log.Printf("%d", len(foo.Reports))
  for _, report := range foo.Reports {
    log.Printf("%s %s %f %f %d\n", report.Id, report.DirTag, report.Lat, report.Lon, report.SecsSinceReport)
  }

  // For now, let's assume that all Trips for a Route have the same Shape
  // N Judah is Route # 1093, Shape 102909
  nReferencer := linref.NewReferencer("102909")
  log.Printf("%f", nReferencer.Reference(37.0,-122.0))
}
