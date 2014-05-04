package main

import (
	"flag"
	"github.com/bdon/go.gtfs"
	"os"
	"os/signal"
	"syscall"
)

var emitFiles bool
var allowAll bool

func init() {
	flag.BoolVar(&emitFiles, "emitFiles", false, "emit files")
	flag.BoolVar(&allowAll, "allowAll", false, "allow all CORS origins")
}

func main() {
	flag.Parse()
	if emitFiles {
		feed := gtfs.Load("muni_gtfs", true)
		EmitStops(feed)
		EmitSchedules(feed)
	} else {
		feed := gtfs.Load("muni_gtfs", false)

		agencyState := NewAgencyState(feed)
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
		Webserver(agencyState, allowAll)
	}
}
