package main

import (
  "path/filepath"
  "fmt"
  "time"
  "strconv"
  "github.com/bdon/jklmnt/nextbus"
  "github.com/bdon/jklmnt/state"
  "encoding/json"
)

func main() {
  const longForm = "2006-01-02 15:04:05 -0700 MST"
  t1, _ := time.Parse(longForm, "2013-08-26 06:00:01 -0700 PDT")
  t2, _ := time.Parse(longForm, "2013-08-27 03:00:01 -0700 PDT")

  list, _ := filepath.Glob("/Volumes/shrub/njudahdata/N/*.xml")
  relevantFiles := []string{}
  for _, entry := range list {
    extension := filepath.Ext(entry)
    filename := filepath.Base(entry)
    var unixstamp = filename[0:len(filename)-len(extension)]
    theint, _ := strconv.ParseInt(unixstamp, 10, 64)
    theTime := time.Unix(theint, 0)
    if theTime.After(t1) && theTime.Before(t2) {
      relevantFiles = append(relevantFiles, entry)
    }
  }

  // now we only have files for 8/26-8/27

  stat := state.NewSystemState()
  for _, entry := range relevantFiles {
    extension := filepath.Ext(entry)
    filename := filepath.Base(entry)
    var unixstamp = filename[0:len(filename)-len(extension)]
    theint, _ := strconv.ParseInt(unixstamp, 10, 64)
    resp := nextbus.ResponseFromFileWithReferencer(entry, stat.Referencer, int(theint))
    stat.AddResponse(resp, int(theint))
  }

  result, _ := json.Marshal(stat.Map)
  fmt.Println(string(result))
}

