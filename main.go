package main

import (
	"flag"
	"github.com/bdon/go.gtfs"
)

var emitFiles bool

func init() {
	flag.BoolVar(&emitFiles, "emitFiles", false, "emit files")
}

func main() {
	flag.Parse()
	if emitFiles {
		feed := gtfs.Load("muni_gtfs", true)
		EmitStops(feed)
		EmitSchedules(feed)
	} else {
		Webserver()
	}
}

