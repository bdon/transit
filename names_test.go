package main

import (
	"testing"
)

func TestName(t *testing.T) {
	byt := []byte(`{"61":"California"}`)
	d := NewNameDict(byt)
	res := d.Resolve("61")

	if res != "California" {
		t.Errorf("Should be California, got %s", res)
	}
}
