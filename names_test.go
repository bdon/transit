package main

import (
	"testing"
)

func TestName(t *testing.T) {
	byt := []byte(`[{"gtfs_short_name":"Powell-Mason","nextbus_name":"59"}]`)
	d := NewNameDict(byt)
	res, _ := d.NtoG("59")

	if res != "Powell-Mason" {
		t.Errorf("Should be Powell-Mason, got %s", res)
	}

	res, _ = d.GtoN("Powell-Mason")

	if res != "59" {
		t.Errorf("Should be 59, got %s", res)
	}

	_, ok := d.GtoN("Nonexistent")
	if ok {
		t.Errorf("OK should be false", res)
	}

	_, ok = d.NtoG("Nonexistent")
	if ok {
		t.Errorf("OK should be false", res)
	}
}
