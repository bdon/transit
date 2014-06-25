package main

import (
	"github.com/bdon/go.gtfs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEmpty(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)
	stat, _ := a.NewRouteState("N")

	if len(stat.Runs) != 0 {
		t.Error("Runs should be empty")
	}
}

func TestLeadingVehicle(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	testResponse := Response{}
	report1 := VehicleReport{LeadingVehicleId: "something", RouteTag: "N"}
	testResponse.Reports = append(testResponse.Reports, report1)
	a.AddResponse(testResponse, 10000000)

	if len(a.RouteStates["N"].Runs) != 0 {
		t.Error("state should ignore reports with vehicle IDs")
	}
}

func TestOne(t *testing.T) {
	feed := gtfs.Load("muni_gtfs", false)
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	testResponse := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
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

	testResponse2 := Response{}
	report2 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
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

	testResponse3 := Response{}
	report3 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
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
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	testResponse := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	testResponse.Reports = append(testResponse.Reports, report1)
	a.AddResponse(testResponse, 10000000)

	testResponse2 := Response{}
	report2 := VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.0",
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
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000000)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := Response{}
	laterReport := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
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
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000000)

	if len(a.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := Response{}
	laterReport := VehicleReport{VehicleId: "1000", DirTag: "OB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	a.AddResponse(laterResponse, 10000001)

	if len(a.RouteStates["N"].Runs) != 2 {
		t.Error("Runs should have 2 elements, because direction changed")
	}
}

func TestSimplify(t *testing.T) {
	run := VehicleRun{VehicleId: "1", Dir: Inbound}
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
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000015)

	response2 := Response{}
	report2 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
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
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)
	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: "", RouteTag: "N"}

	response.Reports = append(response.Reports, report1)
	a.AddResponse(response, 10000015)
	a.Persist(tmpdir)
	b := NewAgencyState(feed, names)
	b.Restore(tmpdir)

	if len(b.RouteStates["N"].Runs) != 1 {
		t.Error("Runs should have one element")
	}

	if (b.RouteStates["N"].Referencer == Referencer{}) {
		t.Error("State should have referencer")
	}

	var curRunN *VehicleRun
	curRunN = b.RouteStates["N"].CurrentRuns["1000"]

	if curRunN != b.RouteStates["N"].Runs["1000_10000000"] {
		t.Error("CurrentRuns should contain a pointer to the first elem")
	}
}

func TestEndDateOfRun(t *testing.T) {
	run := VehicleRun{VehicleId: "1", Dir: Inbound}
	run.States = []VehicleState{}

	state1 := VehicleState{LatString: "0.01", LonString: "0.01", Time: 1402463820}
	state2 := VehicleState{LatString: "0.02", LonString: "0.02", Time: 1402463820}
	state3 := VehicleState{LatString: "0.03", LonString: "0.03", Time: 1402463820}
	run.States = append(run.States, state1)
	run.States = append(run.States, state2)
	run.States = append(run.States, state3)

	y, m, d := run.EndDay()

	if y != 2014 && m != 6 && d != 10 {
		t.Errorf("Expected to get the day")
	}
}

func TestDeleteNotOnToday(t *testing.T) {
	// delete all vehicle reports that are not on Today.
	feed := gtfs.Load("muni_gtfs", false)
	names := NewNameDict([]byte(`[]`))
	a := NewAgencyState(feed, names)
	response1 := Response{}
	response3 := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 0,
		LeadingVehicleId: "", RouteTag: "N"}

	report3 := VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 0,
		LeadingVehicleId: "", RouteTag: "N"}

	response1.Reports = append(response1.Reports, report1)
	response3.Reports = append(response3.Reports, report3)
	a.AddResponse(response1, 1402463820-86401)
	a.AddResponse(response3, 1402463820)

	x := a.DeleteRunsBeforeDay(1402463820)
	if x != 1 {
		t.Errorf("expected to delete 1 run, deleted %d", x)
	}
  x = len(a.RouteStates["N"].Runs)
	if x != 1 {
		t.Errorf("Should only have 1 run total, got %d", x)
	}
}

func TestFilepathForTime(t *testing.T) {
	path := "tmp_path"
	tim, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Sat May 31 14:23:00 -0700 PST 2014")
	p := FilepathForTime(path, tim)
	if p != "tmp_path/2014/5/31" {
		t.Error("Expected path")
	}
}

func TestMkdirpForTime(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "test")
	defer os.Remove(tmpdir)

	tim, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Sat May 31 14:23:00 -0700 PST 2014")
	MkdirpForTime(tmpdir, tim)
	fileinfo, _ := os.Stat(filepath.Join(tmpdir, "2014/5/31"))
	if fileinfo == nil || !fileinfo.IsDir() {
		t.Error("Expected directory")
	}
}
