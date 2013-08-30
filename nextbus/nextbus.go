package nextbus

import (
  "encoding/xml"
  "os"
  "io/ioutil"
  "log"
  "github.com/bdon/jklmnt/linref"
  "strconv"
)

type NextBusVehicleReport struct {
  Id string `xml:"id,attr"`
  DirTag string `xml:"dirTag,attr"`
  LatString string `xml:"lat,attr"`
  LonString string `xml:"lon,attr"`
  SecsSinceReport int `xml:"secsSinceReport,attr"`
  LeadingVehicleId string `xml:"leadingVehicleId,attr"`

  Index float64
  UnixTime int
}

func (report NextBusVehicleReport) Lat() float64 {
  f, _ := strconv.ParseFloat(report.LatString, 64)
  return f
}

func (report NextBusVehicleReport) Lon() float64 {
  f, _ := strconv.ParseFloat(report.LonString, 64)
  return f
}

type NextBusResponse struct {
  Reports []NextBusVehicleReport `xml:"vehicle"`
}

func ResponseFromFileWithReferencer(filename string, r linref.Referencer, unixtime int) NextBusResponse {
  file, err := os.Open(filename)
  if err != nil {
    log.Fatal(err)
  }
  foo := NextBusResponse{}
  body, err := ioutil.ReadAll(file)
  if err != nil {
    log.Fatal(err)
  }
  xml.Unmarshal(body, &foo)

  for i, report := range foo.Reports {
    foo.Reports[i].Index = r.Reference(report.Lat(), report.Lon())
    foo.Reports[i].UnixTime = unixtime - report.SecsSinceReport
  }

  return foo
}
