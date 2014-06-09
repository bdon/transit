package main

import (
	"encoding/json"
)

type Name struct {
	GtfsShortName string `json:"gtfs_short_name"`
	NextbusName   string `json:"nextbus_name"`
}

type NameDict struct {
	Dict []Name
}

func NewNameDict(byt []byte) NameDict {
	var dat []Name
	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	return NameDict{Dict: dat}
}

func (d NameDict) GtoN(g string) (string, bool) {
	for _, x := range d.Dict {
		if x.GtfsShortName == g {
			return x.NextbusName, true
		}
	}
	return "", false
}

func (d NameDict) NtoG(n string) (string, bool) {
	for _, x := range d.Dict {
		if x.NextbusName == n {
			return x.GtfsShortName, true
		}
	}
	return "", false
}
