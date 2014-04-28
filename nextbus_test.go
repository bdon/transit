package transit_timelines

import (
	"testing"
)

func TestDir(t *testing.T) {
	report1 := VehicleReport{DirTag: "N__IB3"}
	report2 := VehicleReport{DirTag: "N__IB1"}
	report3 := VehicleReport{DirTag: "N__OB1"}

	if report1.Dir() != Inbound {
		t.Error("N__IB3 should be Inbound")
	}
	if report2.Dir() != Inbound {
		t.Error("N__IB1 should be Inbound")
	}
	if report3.Dir() != Outbound {
		t.Error("N__OB1 should be Outbound")
	}
}
