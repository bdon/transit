package main

import (
    "fmt"
    "log"
    "encoding/csv"
    "github.com/paulsmith/gogeos/geos"
    "os"
    "io"
    "strconv"
)

func main() {
    file, err := os.Open("muni_gtfs/shapes.txt")
    if err != nil {
        log.Fatal(err)
        return
    }
    defer file.Close()
    reader := csv.NewReader(file)
    reader.TrailingComma = true
    coords := []geos.Coord{}
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        } else if err != nil {
            fmt.Println("Error:", err)
            return
        }
        if record[0] == "102909" {
          lat, _ := strconv.ParseFloat(record[1],64)
          lon, _ := strconv.ParseFloat(record[2],64)
          coords = append(coords, geos.NewCoord(lat, lon))
        }
    }
    // For now, let's assume that all Trips for a Route have the same Shape
    // N Judah is Route # 1093, Shape 102909
    path, _ := geos.NewLineString(coords...)

    for _, coord := range coords {
      point, _ := geos.NewPoint(coord)
      fmt.Println(path.Project(point))
    }
}
