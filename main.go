package main

import (
	"flag"
	"github.com/bdon/go.gtfs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
			log.Println("Done persisting.")
			os.Exit(0)
		}()

		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					x := agencyState.DeleteRunsBeforeDay(int(time.Now().Unix()))
					log.Printf("%d runs deleted.", x)
					agencyState.Persist("static/history")
					log.Println("Done persisting.")
				}
			}
		}()

		agencyState.Start()
		Webserver(agencyState)
	}
}
