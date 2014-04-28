package transit_timelines

import (
	"github.com/bdon/go.gtfs"
	"github.com/paulmach/go.geo"
)

type Referencer struct {
	Path *geo.Path
}

// TODO this should take a geo.Path
func NewReferencer(shapeId string) Referencer {
	ref := Referencer{}

	// Fixme
	feed := gtfs.Load("muni_gtfs")
	route := feed.RouteByShortName("N")
	coords := route.Shapes()[0].Coords
	path := geo.NewPath()
	for _, c := range coords {
		path.Push(geo.NewPoint(c.Lon, c.Lat))
	}
	ref.Path = path
	return ref
}

func (r Referencer) Reference(lat float64, lon float64) float64 {
	point := geo.NewPoint(lon, lat)
	return r.Path.ProjectNormalized(point)
}
