
(function(exports) {
  var Transit = exports.Transit = {};

  Transit.Page = function(static_endpoint, live_endpoint) {
    var my = {};
    var routes = [];
    var calendar = ["1","1","1","1","1","1","1"];
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
      for(r in routes) {
        if(routes[r].id == route.id) return;
      }
      routes.unshift(route);
    }

    my.removeRoute = function(route) {
      dispatch.on("." + route.id, null);
      var newArr = [];
      for(r in routes) {
        if(routes[r].id != route.id) newArr.push(routes[r]);
      }
      routes = newArr;
    }

    my.routes = function() {
      return routes;
    }

    my.setCalendar = function(a) {
      my.calendar = a;
    }

    my.serviceId = function(epoch) {
      var d = new Date(epoch* 1000)
      var day = d.getDay();
      if (day == 0) day = 7;
      return my.calendar[day - 1];
    }

    my.draw = function() {
      var cl = d3.select("#chartlist").selectAll(".route").data(p.routes(), function(d) { return d.id });
      cl.enter().append("div").attr("class", "route").call(timelineChart(p));
      cl.order();
      cl.exit().remove();
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
          var isLive = now - state.time < 120;
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
    
        if(now - lastState.time < 120) {
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

  Transit.Classname = function(prefix, token) {
    return prefix + "_" + token.toLowerCase().replace(/[^a-z0-9]/,"");
  }

  Transit.ShortenName = function(str) {
    if (str.length < 6) return str;
    if (str.indexOf("-") > 0) {
      return str.replace(/([A-Z])[a-z]*/g, function(m) {
        return m[0];
      });
    }
    return str.substring(0,5) + "."; 
  }

})(this);

function timelineChart(p) {
  var routeState = Transit.RouteState();
  var routeSchedule = Transit.RouteSchedule();
  var dir = 0;

  var stopsScale = d3.scale.linear().domain([0,1000]).range([0,150]);
  var axisTimeFormat = d3.time.format("%-I %p");
  var axis = d3.svg.axis().scale(p.timeScale()).orient("top").tickFormat(axisTimeFormat);
  var line = d3.svg.line()
    .x(function(d) { return p.timeScale()(d.time*1000) })
    .y(function(d) { return stopsScale(d.index) })
    .interpolate("linear");

  var vis;
  var clippedFore;
  var clippedBack;
  var timestamp;
  var nowLine;
  var cursor;
  
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

      var control_row = d3.select(this).append("div").attr("class","control_row");
      control_row.append("div").attr("class","route_short_name").text(d.short_name);
      control_row.append("div").attr("class","route_long_name").text(Transit.Camelize(d.long_name));
      control_row.append("div").attr("class","close_button").text("×").on("click", function() {
        p.removeRoute(d);
        p.draw();
        // TODO: hackety hack
        d3.selectAll(".rollsign").filter(function(e) { return e.id == d.id }).classed("displayed", false);
      });

      var contain = control_row.append("div").attr("class","toggle_container");
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

      var format = d3.time.format(":%M");

      rect.on("wheel.zoom",null);
      rect.on("mousewheel.zoom",null);
      rect.on("mousemove.hover", function() {
        var mous = d3.mouse(this);
        var selectedIndex = stopsScale.invert(mous[1]);
        wasSet = false;
        vis.selectAll(".stop").each(function(d,i) {
          if(!wasSet && d.index > selectedIndex) {
            d3.select(this).classed("stopHighlighted",true);
            wasSet = true;
          } else {
            d3.select(this).classed("stopHighlighted",false);
          }
        });

        var selectedTime = p.timeScale().invert(mous[0]);
        var mins = selectedTime.getMinutes();
        if (mins > 4 && mins < 45) {
          cursor.attr("x", mous[0])
            .text(format(selectedTime));
        } else {
          cursor.attr("x",-100);
        }
      });
      mainChart.on("mouseleave", function() {
        cursor.attr("x",-100);
        vis.selectAll(".stop").classed("stopHighlighted",false); 
      });
      
      var clipped = mainChart.append("g")
        .attr("clip-path", "url(#clip_" + d.id + ")")
      clippedBack = clipped.append("g");
      clippedFore = clipped.append("g");

      cursor = mainChartWAxis.append("text")
        .attr("class","timeCursor")
        .attr("y",1);

      nowLine = clippedFore.append("line")
        .attr("class","nowLine")
        .datum(new Date()) 
        .attr({"x1":0,"x2":0,"y1":0,"y2":150});

      d3.json(p.static_endpoint() + "/stops/" + d.id + ".json", function(sch) {
        var dir0 = contain.append("div").attr("class","dir_toggle selected").text(sch.headsigns[0]);
        var dir1 = contain.append("div").attr("class","dir_toggle").text(sch.headsigns[1]);
        contain.on("click", function() {
          dir = (dir == 0 ? 1 : 0);
          dir0.classed("selected",dir == 0);
          dir1.classed("selected",dir == 1);
          bind();
          draw();
        });

        vis.append("g").attr("transform","translate(1024,23)").selectAll(".stop").data(sch.stops).enter().append("text")
            .attr("class", "stop")
            .attr("text-anchor", "begin")
            .attr("y", function(d) { return stopsScale(d.index) })
            .text(function(d) { return d.name });
        bind();
        draw();
      });

      var serviceId = p.serviceId(Transit.Now());
      d3.json(p.static_endpoint() + "/schedules/" + serviceId + "/" + d.id + ".json", function(trips) {
        routeSchedule.parse(trips, Transit.Now());
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
    nowLine.datum(new Date());
    var s1 = clippedBack.selectAll(".guide").data(routeSchedule.trips(dir));
    s1.enter().append("path").attr("class","guide");
    s1.exit().remove();
    var s2 = clippedFore.selectAll(".vehiclePath").data(routeState.trips(now,dir), function(d) { return d.key });
    s2.enter().append("path").attr("class","vehiclePath");
    s2.exit().remove();

    var newdata = routeState.liveVehicles(now,dir);
    var s3 = clippedFore.selectAll(".liveVehicle").data(newdata, function(d) { return d.key });
    var tmp = s3.enter().append("g")
       .attr("class", "liveVehicle")
    tmp.append("polygon").attr("points", "0,4 7,0 0,-4").style("fill","black");
    tmp.append("polygon").attr("points", "0,3 6,0 0,-3")
    s3.exit().remove();
  }

  function draw() {
    nowLine.attr("transform",function(d) { return "translate(" + p.timeScale()(d) + ",0)"})
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
