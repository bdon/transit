package nextbus

import (
	"github.com/bdon/jklmnt/nextbus"
	"testing"
)

func TestDir(t *testing.T) {
	report1 := nextbus.VehicleReport{DirTag: "N__IB3"}
	report2 := nextbus.VehicleReport{DirTag: "N__IB1"}
	report3 := nextbus.VehicleReport{DirTag: "N__OB1"}

	if report1.Dir() != nextbus.Inbound {
		t.Error("N__IB3 should be Inbound")
	}
	if report2.Dir() != nextbus.Inbound {
		t.Error("N__IB1 should be Inbound")
	}
	if report3.Dir() != nextbus.Outbound {
		t.Error("N__OB1 should be Outbound")
	}
}
