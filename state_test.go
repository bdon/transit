package main

import (
	"github.com/bdon/go.gtfs"
	"github.com/bdon/go.nextbus"
	"io/ioutil"
	"os"
	"testing"
)

func TestEmpty(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)
	stat, _ := a.NewRouteState("N")

	if len(stat.Runs) != 0 {
		t.Error("Runs should be empty")
	}
}

func TestLeadingVehicle(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{LeadingVehicleId: "something", RouteTag: "N"}
	testResponse.Reports = append(testResponse.Reports, report1)
	a.AddResponse(testResponse, 10000000)

	if len(a.RouteStates["N"].Runs) != 0 {
		t.Error("state should ignore reports with vehicle IDs")
	}
}

func TestOne(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse.Reports = append(testResponse.Reports, report1)
	a.AddResponse(testResponse, 10000015)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	if len(a.RouteStates["N"].Runs["1000_10000000"].States) != 1 {
		t.Error("First run should have 1 state")
	}

	testResponse2 := nextbus.Response{}
	report2 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	a.AddResponse(testResponse2, 10000015)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
	if len(a.RouteStates["N"].Runs["1000_10000000"].States) != 1 {
		t.Error("First run should have still 1 state if position has not changed")
	}

	testResponse3 := nextbus.Response{}
	report3 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse3.Reports = append(testResponse3.Reports, report3)
	a.AddResponse(testResponse3, 10000015)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
	if len(a.RouteStates["N"].Runs["1000_10000000"].States) != 2 {
		t.Error("First run should have 2 states if position has changed")
	}
}

func TestTwo(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	testResponse := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse.Reports = append(testResponse.Reports, report1)
	a.AddResponse(testResponse, 10000000)

	testResponse2 := nextbus.Response{}
	report2 := nextbus.VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	a.AddResponse(testResponse2, 10000001)

	if len(a.RouteStates["N"].Runs) != 2 {
		t.Error("Runs should have 2 elements")
	}
}

func TestIgnoreFifteenMinutes(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000000)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := nextbus.Response{}
	laterReport := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	a.AddResponse(laterResponse, 10001000)

	if len(a.RouteStates["N"].Runs) != 2 {
		t.Error("Runs should have 2 elements, because too much time passed")
	}
}

func TestChangeDirection(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000000)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := nextbus.Response{}
	laterReport := nextbus.VehicleReport{VehicleId: "1000", DirTag: "OB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	a.AddResponse(laterResponse, 10000001)

	if len(a.RouteStates["N"].Runs) != 2 {
		t.Error("Runs should have 2 elements, because direction changed")
	}
}

func TestSimplify(t *testing.T) {
	run := VehicleRun{VehicleId: "1", Dir: nextbus.Inbound}
	run.States = []VehicleState{}

	state1 := VehicleState{LatString: "0.01", LonString: "0.01"}
	state2 := VehicleState{LatString: "0.02", LonString: "0.02"}
	state3 := VehicleState{LatString: "0.03", LonString: "0.03"}
	run.States = append(run.States, state1)
	run.States = append(run.States, state2)
	run.States = append(run.States, state3)

	//if len(run.States) != 2 {
	//  t.Errorf("States should have 2 elements after simplifying, has %d", len(run.States))
	//}
}

func TestFilteredByTime(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)

	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000015)

	response2 := nextbus.Response{}
	report2 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response2.Reports = append(response2.Reports, report2)
	a.AddResponse(response2, 10000115)

	filtered := a.RouteStates["N"].After(10000099)

	if len(filtered["1000_10000000"].States) != 1 {
		t.Error("Runs should have 1 element")
	}

	if len(a.RouteStates["N"].Runs["1000_10000000"].States) != 2 {
		t.Error("Runs should not have been modified")
	}
}

func TestSaveAndRestore(t *testing.T) {
	// lets write it into a temporary directory
	tmpdir, _ := ioutil.TempDir("", "test")
	defer os.Remove(tmpdir)
	feed := gtfs.Load("muni_gtfs", false)
	a := NewAgencyState(feed)
	response := nextbus.Response{}
	report1 := nextbus.VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000015)
	a.Persist(tmpdir)
	b := NewAgencyState(feed)
	b.Restore(tmpdir)

	if len(b.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have one element")
	}

	// also clears out pointers (plz don't crash)
	if len(b.RouteStates["N"].CurrentRuns) != 1 {
		t.Error("CurrentRuns should only have one element")
	}
}

// test that currentRuns is restored
// test that synchronization primitives added
// test that Shape is added from feed
