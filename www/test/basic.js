describe("Transit", function () {
  fixtures = {};
  fixtures.N = {"id":"1093","short_name":"N","long_name":"Judah"}

  it("Can add a line", function() {
    //tt.addLine({"routeName":"N","routeNum":"1093"});
    var p = Transit.Page();
    p.showRoute(fixtures.N);
    p.showRoute(fixtures.N);
    expect(p.routes()).to.eql([fixtures.N]);
  });

  it("Can delete a line", function() {
    var p = Transit.Page();
    p.removeRoute(fixtures.N);
    p.showRoute(fixtures.N);
    p.removeRoute(fixtures.N);
    expect(p.routes()).to.eql([]);
  });

  it("can add the past data", function() {
    var resp = {
      "1406_1401754326":{
        "vehicle_id":"1406",
        "dir":1,
        "states":[
          {"time":1401754326,"index":454}
        ]}}
    var s = Transit.RouteState();
    s.add(resp);
    expect(s.trips().length).to.eql(1);
    expect(s.trips()[0].key).to.eql("1406_1401754326");
    expect(s.trips()[0].run.states.length).to.eql(1);

    var resp = {
      "1406_1401754326":{
        "vehicle_id":"1406",
        "dir":1,
        "states":[
          {"time":1401754327,"index":455}
        ]}}

    s.add(resp);
    expect(s.trips().length).to.eql(1);
    var states = s.trips()[0].run.states;
    expect(states.length).to.eql(2);
    expect(states[0].time).to.eql(1401754326);
    expect(states[0].index).to.eql(454);
    expect(states[1].time).to.eql(1401754327);
    expect(states[1].index).to.eql(455);
  });

  it("can filter trips by direction", function() {
    var resp = {
      "1406_1401754326":{
        "vehicle_id":"1406",
        "dir":1,
        "states":[
          {"time":1401754327,"index":455}
        ]
      },
      "1406_1401754999":{
        "vehicle_id":"1406",
        "dir":0,
        "states":[
          {"time":1401754327,"index":455}
        ]
      }
    }

    var s = Transit.RouteState();
    s.add(resp);
    expect(s.trips(1).length).to.eql(1);
    expect(s.trips(1)[0].run.dir).to.eql(1);
  });

  it("normalizes schedule times to current day", function() {
    var response = [
      {
        "trip_id":"5681922",
        "dir":"0",
        "stops":[
          {"time":10,"index":27},
          {"time":20,"index":64}
        ]
      },
      {
        "trip_id":"5681923",
        "dir":"0",
        "stops":[
          {"time":10,"index":27},
          {"time":20,"index":64}
        ]
      }
    ];

    var s = Transit.RouteSchedule();

    var simulatedEpoch = 1401774229;
    s.parse(response, simulatedEpoch);

    var midnight = Transit.Midnight(simulatedEpoch);
    expect(s.trips().length).to.eql(2);
    expect(s.trips()[0].stops[0].time).to.eql(midnight+10);
    expect(s.trips()[0].dir).to.be.a("number"); //TODO: bleh.
  });

  // i don't know if this is legit. what about 24/7 routes.
  it("normalizes to the previous day if we're before first day departure", function() {
    var response = [
      {
        "trip_id":"5681922",
        "dir":"0",
        "stops":[
          {"time":10,"index":27},
          {"time":20,"index":64}
        ]
      }
    ];

    var s = Transit.RouteSchedule();

    var arbitrary = 1401774229;
    var midnight = Transit.Midnight(arbitrary);
    var simulatedEpoch = midnight + 5;

    s.parse(response, simulatedEpoch);

    expect(s.trips()[0].stops[0].time).to.eql(midnight+10-86400);
  });

  it("calculates midnight before a given epoch time", function() {
    // this is in local time (PDT)
    var m = Transit.Midnight(1401774229);
    expect(m).to.eql(1401692400);
  });

  it("can filter schedules by direction", function() {
    var response = [
      {
        "trip_id":"5681922",
        "dir":"0",
        "stops":[
          {"time":10,"index":27},
          {"time":20,"index":64}
        ]
      },
      {
        "trip_id":"5681922",
        "dir":"1",
        "stops":[
          {"time":10,"index":27},
          {"time":20,"index":64}
        ]
      }
    ];

    var s = Transit.RouteSchedule();

    var arbitrary = 1401774229;
    var midnight = Transit.Midnight(arbitrary);
    s.parse(response, arbitrary);

    expect(s.trips(1).length).to.eql(1);
    expect(s.trips(1)[0].dir).to.eql(1);
  });
});

// it requests for static/ if the chosen date is not today
// if requests for live endpoint if it is

// TODO: enforce order of timestamps in a route state
