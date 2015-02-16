package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type Direction int

const (
	Outbound Direction = iota
	Inbound  Direction = iota
)

type VehicleReport struct {
	VehicleId        string `xml:"id,attr"`
	DirTag           string `xml:"dirTag,attr"`
	LatString        string `xml:"lat,attr"`
	LonString        string `xml:"lon,attr"`
	SecsSinceReport  int    `xml:"secsSinceReport,attr"`
	LeadingVehicleId string `xml:"leadingVehicleId,attr"`
	RouteTag         string `xml:"routeTag,attr"`

	UnixTime int
}

func (report VehicleReport) Lat() float64 {
	f, _ := strconv.ParseFloat(report.LatString, 64)
	return f
}

func (report VehicleReport) Lon() float64 {
	f, _ := strconv.ParseFloat(report.LonString, 64)
	return f
}

func (report VehicleReport) Dir() Direction {
	if strings.Contains(report.DirTag, "I_") {
		return Inbound
	} else {
		return Outbound
	}
}

type Response struct {
	Reports []VehicleReport `xml:"vehicle"`
}

func ResponseFromFile(filename string, unixtime int) Response {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	response := Response{}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	xml.Unmarshal(body, &response)

	for i, report := range response.Reports {
		response.Reports[i].UnixTime = unixtime - report.SecsSinceReport
	}

	return response
}
