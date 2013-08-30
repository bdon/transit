package main

import (
	"encoding/json"
	"fmt"
	"github.com/bdon/jklmnt/nextbus"
	"github.com/bdon/jklmnt/state"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	const longForm = "2006-01-02 15:04:05 -0700 MST"
	t1, _ := time.Parse(longForm, "2013-08-26 06:00:01 -0700 PDT")
	t2, _ := time.Parse(longForm, "2013-08-27 03:00:01 -0700 PDT")

	list, _ := filepath.Glob("/Users/Bdon/workspace/njudahdata/N/*.xml")
	relevantFiles := []string{}
	for _, entry := range list {
		unixstamp := filepathToTime(entry)
		theTime := time.Unix(unixstamp, 0)
		if theTime.After(t1) && theTime.Before(t2) {
			relevantFiles = append(relevantFiles, entry)
		}
	}

	// now we only have files for 8/26-8/27

	stat := state.NewSystemState()
	for _, entry := range relevantFiles {
		unixstamp := filepathToTime(entry)
		resp := nextbus.ResponseFromFile(entry, int(unixstamp))
		stat.AddResponse(resp, int(unixstamp))
	}

	result, _ := json.Marshal(stat.Runs)
	fmt.Println(string(result))
}

func filepathToTime(entry string) int64 {
	extension := filepath.Ext(entry)
	filename := filepath.Base(entry)
	var unixstamp = filename[0 : len(filename)-len(extension)]
	theint, _ := strconv.ParseInt(unixstamp, 10, 64)
	return theint
}
