package main

import (
	"flag"
	"github.com/bdon/go.gtfs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var emitFiles bool

func init() {
	flag.BoolVar(&emitFiles, "emitFiles", false, "emit files")
}

func main() {
	flag.Parse()

	desc, _ := ioutil.ReadFile("names.json")
	names := NewNameDict(desc)

	if emitFiles {
		feed := gtfs.Load("muni_gtfs", true)
		EmitSchedules(feed)
		EmitStops(feed)
		EmitRoot(feed)
	} else {
		feed := gtfs.Load("muni_gtfs", false)

		agencyState := NewAgencyState(feed, names)
		agencyState.Restore("static/history")

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)

		go func() {
			<-c
			agencyState.Persist("static/history")
			os.Exit(0)
		}()

		agencyState.Start()
		Webserver(agencyState)
	}
}
