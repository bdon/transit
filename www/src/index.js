(function(exports) {
  var Transit = exports.Transit = {};

  Transit.Page = function(static_endpoint, live_endpoint) {
    var my = {};
    var routes = {};
    var timeScale = d3.time.scale().range([0,1020]);
    var dispatch = d3.dispatch("zoom","update");
    var zoom = d3.behavior.zoom().scaleExtent([1,4]).on("zoom", function() { dispatch.zoom(); });
    var prevMidnight = (new Date()).setHours(0,0,0,0);
    var nextMidnight = (new Date()).setHours(24,0,0,0);

    timeScale.domain([prevMidnight, nextMidnight]);
    zoom.x(timeScale);
    zoom.scale(1);

    my.showRoute = function(route) {
      console.log("Showing ", route);
      routes[route.id] = route;
    }

    my.removeRoute = function(route) {
      console.log("Removing ", route);
      dispatch.on("." + route.id, null);
      delete routes[route.id];
    }

    my.routes = function() {
      vals = [];
      for(var k in routes) vals.push(routes[k]);
      return vals;
    }

    my.zoom = function() { return zoom; }
    my.timeScale = function() { return timeScale; }
    my.dispatch = function() { return dispatch; }
    my.update = function() { dispatch.update(); }
    my.static_endpoint = function() { return static_endpoint; }
    my.live_endpoint = function() { return live_endpoint; }

    return my;
  }

  Transit.RouteState = function() {
    var my = {};
    var trips = {};

    my.add = function(resp) {
      for (var k in resp) {
        if(trips[k]) {
          for (var s in resp[k].states) {
            trips[k].states.push(resp[k].states[s]);
          }
        } else {
          trips[k] = resp[k];
        } 
      }
    }

    my.trips = function() { 
      var retval = [];
      for (k in trips) {
        retval.push({"key":k,"run":trips[k]});
      }
      return retval;
    };

    return my;
  }

  Transit.RouteSchedule = function() {
    var my = {};
    var trips = [];

    my.parse = function(response, now) {
      var min = d3.min(response, function(trip) {
        return d3.min(trip.stops, function(stop) { return stop.time })
      })

      var midnightUnix = Transit.Midnight(now);

      if (now - midnightUnix < min) {
        midnightUnix = midnightUnix - 86400;
      }

      for (var trip in response) {
        for (var stop in response[trip].stops) {
          response[trip].stops[stop].time = response[trip].stops[stop].time + midnightUnix;
        }
      }
      trips = response;
    }

    my.trips = function() {
      return trips;
    }

    return my;
  }

  Transit.Midnight = function(epoch) {
    var date = new Date(epoch * 1000);
    date.setHours(0,0,0,0);
    return date.getTime() / 1000;
  }

})(this);

function timelineChart(p) {
  var routeState = Transit.RouteState();
  var routeSchedule = Transit.RouteSchedule();
  var inbound = true;

  var stopsScale = d3.scale.linear().domain([0,1000]).range([0,150]);
  var axis = d3.svg.axis().scale(p.timeScale()).orient("top");
  var line = d3.svg.line()
    .x(function(d) { return p.timeScale()(d.time*1000) })
    .y(function(d) { return stopsScale(d.index) })
    .interpolate("linear");

  var vis;
  var clippedFore;
  var nextbus_route;
  var timestamp;
  
  function my(selection) {
    // this only handles a single one...
    selection.each(function(d, i) {
      console.log("Registering with the Page.")
      p.dispatch().on("zoom." + d.id,draw);
      timestamp = new Date().getTime();
      p.dispatch().on("update." + d.id, function() {
        getDataSince(timestamp)
        timestamp = new Date().getTime();
      });
      
      nextbus_route = d.short_name;
      d3.select(this).append("div").attr("class","nextbus_route").text(nextbus_route + " " + d.long_name);

      var svg = d3.select(this).append("svg:svg")
        .attr("width","100%")
        .attr("height","200px")
        .attr("class", "muni")
        .attr("viewBox","0 0 1200 200")
        .attr("preserveAspectRatio","xMaxYMid slice");
      vis = svg.append("svg:g")
          .attr("transform", "translate(16,16)");

      var switchG = vis.append("g").attr("transform","translate(1040,0)");
      switchG.append("rect").attr("width",100).attr("height",20).style("fill","#aaa")
        .on("click", function(d) {
          inbound = !inbound;
          draw();
        });
      switchG.append("text").text("Switch").attr("y",15);

      var mainChartWAxis = vis.append("g").attr("transform","translate(0,0)")
      mainChartWAxis.append("g")
        .attr("transform", "translate(0,10)")
        .attr("class", "time axis")
      var mainChart = mainChartWAxis.append("g").attr("transform","translate(0,20)")
      mainChart.append("clipPath")
        .attr("id", "clip_" + d.id)
        .append("rect")
          .attr("width","1020")
          .attr("height","550");
      mainChart.append("rect")
        .attr("class", "pane")
        .attr("width","1020")
        .attr("height","550")
        .call(p.zoom());
      var clipped = mainChart.append("g")
        .attr("clip-path", "url(#clip_" + d.id + ")")
      var clippedBack = clipped.append("g");
      clippedFore = clipped.append("g");

      d3.json(p.static_endpoint() + "/stops/" + d.id + ".json", function(stops) {
        vis.append("g").attr("transform","translate(1024,23)").selectAll(".stop").data(stops).enter().append("text")
            .attr("class", "stop")
            .attr("text-anchor", "begin")
            .attr("y", function(d) { return stopsScale(d.index) })
            .text(function(d) { return d.name });
        draw();
      });

      d3.json(p.static_endpoint() + "/schedules/" + d.id + ".json", function(trips) {
        var now = (new Date()).getTime() / 1000;
        routeSchedule.parse(trips, now);
        clippedBack.selectAll(".guide").data(routeSchedule.trips()).enter().append("path")
          .attr("class","guide");
        draw();
      });

      getPastData();
    }); // selection.each
  }

  function draw() {
    clippedFore.selectAll(".vehiclePath").data(routeState.trips(), function(d) { return d.key }).enter()
      .append("path")
      .attr("class","vehiclePath");
    vis.selectAll(".vehiclePath").attr("d", function(d) { return line(d.run.states); })
    vis.selectAll(".guide").attr("d", function(d) { return line(d.stops) });
    vis.selectAll(".time.axis").call(axis);
  }

  function getPastData() {
    d3.json(p.live_endpoint() + "/locations.json?route=" + nextbus_route, function(response) {
      routeState.add(response);
      draw();
    });
  }

  function getDataSince(timestamp) {
    d3.json(p.live_endpoint() + "/locations.json?route=" + nextbus_route + "&after=" + Math.floor(timestamp/1000), function(response) {
      routeState.add(response);
      draw();
    });
  }

  return my;
}
