(function(exports) {
  var Transit = exports.Transit = {};

  Transit.Page = function(static_endpoint, live_endpoint) {
    var my = {};
    var routes = {};
    var timeScale = d3.time.scale().range([0,1020]);
    var dispatch = d3.dispatch("zoom","update");
    var zoom = d3.behavior.zoom().scaleExtent([3,3]).on("zoom", function() { dispatch.zoom(); })
    var prevMidnight = (new Date()).setHours(0,0,0,0);
    var nextMidnight = (new Date()).setHours(24,0,0,0);

    timeScale.domain([prevMidnight, nextMidnight]);
    zoom.x(timeScale);
    zoom.scale(3);
    var trans = timeScale((new Date).getTime()) - 900;
    zoom.translate([-trans,0]);

    my.showRoute = function(route) {
      routes[route.id] = route;
    }

    my.removeRoute = function(route) {
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

    my.trips = function(now, dir) { 
      var retval = [];
      for (k in trips) {
        if(arguments.length < 2 || trips[k].dir == dir) {
          var states = trips[k].states;
          var state = states[states.length-1];
          var isLive = now - state.time < 60;
          retval.push({"key":k,"run":trips[k],"isLive":isLive});
        }
      }
      return retval;
    };

    my.liveVehicles = function(now, dir) {
      retval = [];
      for (k in trips) {
        var states = trips[k].states;
        if (states.length < 2) continue;
        var lastState = states[states.length-1];
        var beforeState = states[states.length-2];
    
        if(now - lastState.time < 60) {
          if(arguments.length == 1 || trips[k].dir == dir) {
            if(lastState.index > 50 && lastState.index < 950) {
              retval.push({
                "time":lastState.time,
                "index":lastState.index,
                "key":k,
                "prev":{"time":beforeState.time,"index":beforeState.index}
              });
            }
          }
        }
      }
      return retval;
    }

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
          response[trip].dir = +response[trip].dir;
        }
      }
      trips = response;
    }

    my.trips = function(dir) {
      if (arguments.length) {
        return trips.filter(function(t) { return t.dir == dir; });
      }
      return trips;
    }

    return my;
  }

  Transit.Midnight = function(epoch) {
    var date = new Date(epoch * 1000);
    date.setHours(0,0,0,0);
    return date.getTime() / 1000;
  }

  Transit.Now = function() {
    return Math.floor((new Date()).getTime() / 1000);
  }

  Transit.Camelize = function(input) {
    return input.toLowerCase().replace(/(\b|-)\w/g,function(m) {
      return m.toUpperCase();
    });
  }

})(this);

function timelineChart(p) {
  var routeState = Transit.RouteState();
  var routeSchedule = Transit.RouteSchedule();
  var dir = 1;

  var stopsScale = d3.scale.linear().domain([0,1000]).range([0,150]);
  var axis = d3.svg.axis().scale(p.timeScale()).orient("top");
  var line = d3.svg.line()
    .x(function(d) { return p.timeScale()(d.time*1000) })
    .y(function(d) { return stopsScale(d.index) })
    .interpolate("linear");

  var vis;
  var clippedFore;
  var clippedBack;
  var timestamp;
  
  function my(selection) {
    // this only handles a single one...
    selection.each(function(d, i) {

      p.dispatch().on("zoom." + d.id,draw);

      timestamp = new Date().getTime();
      p.dispatch().on("update." + d.id, function() {
        d3.json(p.live_endpoint() + "/locations.json?route=" + d.short_name + "&after=" + Math.floor(timestamp/1000), function(response) {
          routeState.add(response);
          bind();
          draw();
        });
        timestamp = new Date().getTime();
      });
      
      d3.select(this).append("div").attr("class","route_short_name").text(d.short_name);
      d3.select(this).append("div").attr("class","route_long_name").text(Transit.Camelize(d.long_name));
      d3.select(this).append("div").attr("class","close_button").text("Close");
      d3.select(this).append("div").attr("class","switch_button").text("Switch").on("click", function(d) {
          if (dir == 0) dir = 1;
          else dir = 0;
          bind();
          draw();
      });

      var svg = d3.select(this).append("svg:svg")
        .attr("width","100%")
        .attr("height","200px")
        .attr("class", "muni")
        .attr("viewBox","0 0 1200 200")
        .attr("preserveAspectRatio","xMaxYMid slice");
      vis = svg.append("svg:g")
          .attr("transform", "translate(16,16)");

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
      var rect = mainChart.append("rect")
        .attr("class", "pane")
        .attr("width","1020")
        .attr("height","550")
        .call(p.zoom());

      rect.on("wheel.zoom",null);
      rect.on("mousewheel.zoom",null);
      rect.on("mousemove.hover", function() {
        console.log("Hovering");
      });
      
      var clipped = mainChart.append("g")
        .attr("clip-path", "url(#clip_" + d.id + ")")
      clippedBack = clipped.append("g");
      clippedFore = clipped.append("g");

      d3.json(p.static_endpoint() + "/schedules/" + d.id + ".json", function(trips) {
        routeSchedule.parse(trips.trips, Transit.Now());
        vis.append("g").attr("transform","translate(1024,23)").selectAll(".stop").data(trips.stops).enter().append("text")
            .attr("class", "stop")
            .attr("text-anchor", "begin")
            .attr("y", function(d) { return stopsScale(d.index) })
            .text(function(d) { return d.name });
        bind();
        draw();
      });

      d3.json(p.live_endpoint() + "/locations.json?route=" + d.short_name , function(response) {
        routeState.add(response);
        bind();
        draw();
      });
    }); // selection.each
  }

  function bind() {
    now = Transit.Now();
    var s1 = clippedBack.selectAll(".guide").data(routeSchedule.trips(dir));
    s1.enter().append("path").attr("class","guide");
    s1.exit().remove();
    var s2 = clippedFore.selectAll(".vehiclePath").data(routeState.trips(now,dir), function(d) { return d.key });
    s2.enter().append("path").attr("class","vehiclePath");
    s2.exit().remove();
    var newdata = routeState.liveVehicles(now,dir);
    var s3 = clippedFore.selectAll(".liveVehicle").data(newdata, function(d) { return d.key });
    s3.enter().append("polygon")
       .attr("class", "liveVehicle")
       .attr("points", "0,3 6,0 0,-3")
    s3.exit().remove();
  }

  function draw() {
    vis.selectAll(".vehiclePath")
      .attr("d", function(d) { return line(d.run.states); })
      .classed("live", function(d) { return d.isLive });
    vis.selectAll(".guide").attr("d", function(d) { return line(d.stops) });
    vis.selectAll(".liveVehicle")
      .attr("transform", function(d) {
        var x2 = p.timeScale()(d.time*1000);
        var y2 = stopsScale(d.index);
        var x1 = p.timeScale()(d.prev.time*1000);
        var y1 = stopsScale(d.prev.index);
        var angle = Math.atan2(-(y2-y1),x2-x1)*180/Math.PI;
        return "translate(" + x1 + "," + y1 + ") rotate(" + -angle + ")";
      });

    vis.selectAll(".time.axis").call(axis);
  }

  return my;
}
