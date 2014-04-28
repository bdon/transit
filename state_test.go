package transit_timelines

import (
	"testing"
)

func TestEmpty(t *testing.T) {
	stat := NewSystemState()

	if len(stat.Runs) != 0 {
		t.Error("Runs should be empty")
	}
}

func TestLeadingVehicle(t *testing.T) {
	stat := NewSystemState()

	testResponse := Response{}
	report1 := VehicleReport{LeadingVehicleId: "something"}
	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000000)

	if len(stat.Runs) != 0 {
		t.Error("state should ignore reports with vehicle IDs")
	}
}

func TestOne(t *testing.T) {
	stat := NewSystemState()

	testResponse := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000015)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	if len(stat.Runs["1000_10000000"].States) != 1 {
		t.Error("First run should have 1 state")
	}

	testResponse2 := Response{}
	report2 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	stat.AddResponse(testResponse2, 10000015)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
	if len(stat.Runs["1000_10000000"].States) != 1 {
		t.Error("First run should have still 1 state if position has not changed")
	}

	testResponse3 := Response{}
	report3 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse3.Reports = append(testResponse3.Reports, report3)
	stat.AddResponse(testResponse3, 10000015)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}
	if len(stat.Runs["1000_10000000"].States) != 2 {
		t.Error("First run should have 2 states if position has changed")
	}
}

func TestTwo(t *testing.T) {
	stat := NewSystemState()

	testResponse := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse.Reports = append(testResponse.Reports, report1)
	stat.AddResponse(testResponse, 10000000)

	testResponse2 := Response{}
	report2 := VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	testResponse2.Reports = append(testResponse2.Reports, report2)
	stat.AddResponse(testResponse2, 10000001)

	if len(stat.Runs) != 2 {
		t.Error("Runs should have 2 elements")
	}
}

func TestIgnoreFifteenMinutes(t *testing.T) {
	stat := NewSystemState()

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000000)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := Response{}
	laterReport := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	stat.AddResponse(laterResponse, 10001000)

	if len(stat.Runs) != 2 {
		t.Error("Runs should have 2 elements, because too much time passed")
	}
}

func TestChangeDirection(t *testing.T) {
	stat := NewSystemState()

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000000)

	if len(stat.Runs) != 1 {
		t.Error("Runs should have 1 element")
	}

	laterResponse := Response{}
	laterReport := VehicleReport{VehicleId: "1000", DirTag: "OB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	laterResponse.Reports = append(laterResponse.Reports, laterReport)
	stat.AddResponse(laterResponse, 10000001)

	if len(stat.Runs) != 2 {
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

	run.Simplify()
	//if len(run.States) != 2 {
	//  t.Errorf("States should have 2 elements after simplifying, has %d", len(run.States))
	//}
}

func TestFilteredByTime(t *testing.T) {
	stat := NewSystemState()

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000015)

	response2 := Response{}
	report2 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response2.Reports = append(response2.Reports, report2)
	stat.AddResponse(response2, 10000115)

	filtered := stat.After(10000099)

	if len(filtered["1000_10000000"].States) != 1 {
		t.Error("Runs should have 1 element")
	}

	if len(stat.Runs["1000_10000000"].States) != 2 {
		t.Error("Runs should not have been modified")
	}
}

// delete runs that started more than 12 hours ago
func TestDeleteOlderThan(t *testing.T) {
	stat := NewSystemState()

	response := Response{}
	report1 := VehicleReport{VehicleId: "1000", DirTag: "IB", LatString: "37.0",
		LonString: "-122.0", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response.Reports = append(response.Reports, report1)
	stat.AddResponse(response, 10000015)

	response2 := Response{}
	report2 := VehicleReport{VehicleId: "1001", DirTag: "IB", LatString: "37.1",
		LonString: "-122.1", SecsSinceReport: 15,
		LeadingVehicleId: ""}

	response2.Reports = append(response2.Reports, report2)
	stat.AddResponse(response2, 10000115)

	stat.DeleteOlderThan(60*60, 10000000+60*61)
	if len(stat.Runs) != 1 {
		t.Error("Runs should only have one element")
	}

	// also clears out pointers (plz don't crash)
	if len(stat.CurrentRuns) != 1 {
		t.Error("CurrentRuns should only have one element")
	}
}
