package main

import (
	"encoding/json"
)

type NameDict struct {
	Dict map[string]string
}

func NewNameDict(byt []byte) NameDict {
	var dat map[string]string
	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	return NameDict{Dict: dat}
}

func (d NameDict) Resolve(n string) string {
	return d.Dict[n]
}
