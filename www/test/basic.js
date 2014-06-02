describe("Control.Attribution", function () {

  it("should foo", function() {
    expect(true).to.eql(true)
  });

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
});

// mock data
[{}]

// mock update data

// request for multiple lines

// it requests for static/ if the chosen date is not today
// if requests for live endpoint if it is
