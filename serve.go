package main

import (
  "fmt"
  "net/http"
  "github.com/bdon/jklmnt/nextbus"
  "encoding/json"
)


type Thingy struct {
  Response nextbus.NextBusResponse
}

func (t Thingy) handler(w http.ResponseWriter, r *http.Request) {
    result, _ := json.Marshal(t.Response.Repr())
    //fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    fmt.Fprintf(w, string(result))
}

func main() {
  thingy := Thingy{}
  thingy.Response = nextbus.ResponseFromFile("N/1377452461.xml")
  http.HandleFunc("/", thingy.handler)
  http.ListenAndServe(":8080", nil)
}
