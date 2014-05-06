package main

import (
	"github.com/bdon/go.gtfs"
	"github.com/paulmach/go.geo"
)

type Referencer struct {
	Path *geo.Path
}

func NewReferencer(coords []gtfs.Coord) Referencer {
	ref := Referencer{}
	path := geo.NewPath()
	for _, c := range coords {
		path.Push(geo.NewPoint(c.Lon, c.Lat))
	}
	ref.Path = path
	return ref
}

func (r Referencer) Reference(lat float64, lon float64) int {
	point := geo.NewPoint(lon, lat)
	return int(r.Path.Project(point)*1000 + 0.5)
}
