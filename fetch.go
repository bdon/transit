package main

import (
  "os"
  "net/http"
  "io"
  "time"
  "fmt"
)

func main() {
  _ = os.Mkdir("N", 0777)
  var now = time.Now()
  out, _ := os.Create(fmt.Sprintf("N/%d.xml", now.Unix()))
  defer out.Close()
  resp, _ := http.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=vehicleLocations&a=sf-muni&r=N&t=0")
  defer resp.Body.Close()
  _, _ = io.Copy(out, resp.Body)
}
