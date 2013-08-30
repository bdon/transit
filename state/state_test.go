package state

import (
	"github.com/bdon/jklmnt/nextbus"
	"github.com/bdon/jklmnt/state"
	"testing"
)

func TestEmpty(t *testing.T) {
	stat := state.NewSystemState()

	if len(stat.Runs) != 0 {
		t.Error("Runs should be empty")
	}
}

func TestLeadingVehicle(t *testing.T) {
	stat := state.NewSystemState()

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{LeadingVehicleId: "something"}
	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000000)

	if len(stat.Runs) != 0 {
		t.Error("state should ignore reports with vehicle IDs")
	}
}

func TestOne(t *testing.T) {
	stat := state.NewSystemState()

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000000)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
  if len(stat.Runs[0].States) != 1 {
    t.Error("First run should have 1 state")
  }

	testResponse2 := nextbus.Response{}
	report2 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	stat.AddResponse(testResponse2, 10000001)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
  if len(stat.Runs[0].States) != 1 {
    t.Error("First run should have still 1 state if position has not changed")
  }

	testResponse3 := nextbus.Response{}
	report3 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse3.Reports = append(testResponse3.Reports, report3)
	stat.AddResponse(testResponse3, 10000002)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
  if len(stat.Runs[0].States) != 2 {
    t.Error("First run should have 2 states if position has changed")
  }
}

func TestTwo(t *testing.T) {
	stat := state.NewSystemState()

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000000)

	testResponse2 := nextbus.Response{}
	report2 := nextbus.VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	stat.AddResponse(testResponse2, 10000001)

	if len(stat.Runs) != 2 {
		t.Error("Runs should have 2 elements")
	}
}

func TestIgnoreFifteenMinutes(t *testing.T) {
	stat := state.NewSystemState()

	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000000)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := nextbus.Response{}
	laterReport := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	stat.AddResponse(laterResponse, 10001000)

	if len(stat.Runs) != 2 {
		t.Error("Runs should have 2 elements, because too much time passed")
	}
}

func TestChangeDirection(t *testing.T) {
	stat := state.NewSystemState()

	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000000)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := nextbus.Response{}
	laterReport := nextbus.VehicleReport{VehicleId: "1000", DirTag: "OB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	stat.AddResponse(laterResponse, 10000001)

	if len(stat.Runs) != 2 {
		t.Error("Runs should have 2 elements, because direction changed")
	}
}
