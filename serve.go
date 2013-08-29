package main

import (
  "net/http"
  "log"
  "time"
  "github.com/bdon/jklmnt/state"
)

func main() {
  s := state.NewSystemState()
  ticker := time.NewTicker(10 * time.Second)
  go func() {
    for {
      select {
        case t := <-ticker.C:
          s.Tick(int(t.Unix()))
      }
    }
  }()

  // do the initial thing
  go s.Tick(int(time.Now().Unix()))

  http.HandleFunc("/", s.Handler)
  log.Println("Serving on port 8080.")
  http.ListenAndServe(":8080", nil)
}
