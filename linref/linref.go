package linref

import (
    "fmt"
    "log"
    "encoding/csv"
    "github.com/paulsmith/gogeos/geos"
    "os"
    "io"
    "strconv"
)

type Referencer struct {
  Path *geos.Geometry
}

func NewReferencer(shapeId string) Referencer {
  ref := Referencer{}
  // Fixme
  file, err := os.Open("/Users/Bdon/workspace/gopath/src/github.com/bdon/jklmnt/muni_gtfs/shapes.txt")
  if err != nil {
      log.Fatal(err)
      return ref
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
          return ref
      }
      if record[0] == shapeId {
        lon, _ := strconv.ParseFloat(record[1],64)
        lat, _ := strconv.ParseFloat(record[2],64)
        coords = append(coords, geos.NewCoord(lat, lon))
      }
  }
  path, _ := geos.NewLineString(coords...)
  ref.Path = path
  return ref
}

func (r Referencer) Reference(lat float64, lon float64) float64 {
  coord := geos.NewCoord(lat, lon)
  point, _ := geos.NewPoint(coord)
  return r.Path.ProjectNormalized(point)
}
