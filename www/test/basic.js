describe("Control.Attribution", function () {

  var tt;
	beforeEach(function () {
	  tt = ttMain("http://static.example.com","http://live.example.com")
	});

  it("should foo", function() {
    expect(true).to.eql(true)
  });

  it("Can add a line", function() {

    tt.addLine({"routeName":"N","routeNum":"1093"});


   
  });
});

// mock data
[{}]

// mock update data

// request for multiple lines

// it requests for static/ if the chosen date is not today
// if requests for live endpoint if it is
