package nextbus

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

  Index float64
  UnixTime int
}

type NextBusVehicleReportRepr struct {
  Vid string `json:"vid"`
  Index float64 `json:"index"`
  Time int `json:"time"`
}

type NextBusResponseRepr struct {
  Reports []NextBusVehicleReportRepr
}

type NextBusResponse struct {
  Reports []NextBusVehicleReport `xml:"vehicle"`
}

func (response NextBusResponse) Repr() NextBusResponseRepr {
  retval := NextBusResponseRepr{}
  reprList := []NextBusVehicleReportRepr{}
  for _, report := range response.Reports {
    newReport := NextBusVehicleReportRepr{}
    newReport.Index = report.Index
    newReport.Time = report.UnixTime
    newReport.Vid = report.Id
    reprList = append(reprList, newReport)
  }
  retval.Reports = reprList
  return retval
}

func ResponseFromFile(filename string) NextBusResponse {
  // For now, let's assume that all Trips for a Route have the same Shape
  // N Judah is Route # 1093, Shape 102909
  nReferencer := linref.NewReferencer("102909")

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
    foo.Reports[i].Index = nReferencer.Reference(report.Lat, report.Lon)
    foo.Reports[i].UnixTime = 1377452461 - report.SecsSinceReport
  }

  return foo
}

