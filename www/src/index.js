var STATIC_ENDPOINT = "http://localhost:8081/static";
var LIVE_ENDPOINT = "http://localhost:8080";

// The main object for the page.
// Controls timers, scales, lines.
function ttMain(staticEndpoint, liveEndpoint) {

  var charts = [];
  function drawAllCharts() {
    for (var i in charts) { charts[i].d(); }
  }

  var staticE = staticEndpoint;
  var liveE = liveEndpoint;
  var timeScale = d3.time.scale().range([0,1020]);
  var zoom = d3.behavior.zoom().scaleExtent([1,4]).on("zoom", drawAllCharts);

  d = new Date();
  var prevMidnight = d.setHours(0,0,0,0);

  d = new Date();
  var nextMidnight = d.setHours(24,0,0,0);

  timeScale.domain([prevMidnight, nextMidnight]);
  zoom.x(timeScale);
  zoom.scale(1);

  var my = {};

  my.zoom = function() {
    if(!arguments.length) return zoom;
  }

  my.timeScale = function() {
    if(!arguments.length) return timeScale;
  }

  my.addChart = function(c) {
    charts.push(c); 
  }
  
  return my;
}

function timelineChart(z,ts) {
  var lastTime;
  var timeScale = ts;
  var stopsScale = d3.scale.linear().domain([0,1000]).range([0,150]);
  var axis = d3.svg.axis().scale(timeScale).orient("top")
  var data = [];
  var ends = [];
  var line = d3.svg.line()
    .x(function(d) { return timeScale(d.time*1000) })
    .y(function(d) { return stopsScale(d.index) })
    .interpolate("linear");
  var zoom = z;
  var inbound = true;

  // ???
  var vis;
  var clippedFore;
  var nextbus_route;

  function my(selection) {
    selection.each(function(d, i) {
      nextbus_route = d.nextbus_route;

      d3.select(this).append("div").attr("class","nextbus_route").text(nextbus_route);

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
          vis.selectAll(".inbound").classed("hidden",!inbound);
          vis.selectAll(".outbound").classed("hidden",inbound);
        });
      switchG.append("text").text("Switch").attr("y",15);

      var mainChartWAxis = vis.append("g").attr("transform","translate(0,0)")
      mainChartWAxis.append("g")
        .attr("transform", "translate(0,10)")
        .attr("class", "time axis")

      var mainChart = mainChartWAxis.append("g").attr("transform","translate(0,20)")

      mainChart.append("clipPath")
        .attr("id", "clip")
        .append("rect")
          .attr("width","1020")
          .attr("height","550");

      mainChart.append("rect")
        .attr("class", "pane")
        .attr("width","1020")
        .attr("height","550")
        .call(zoom);

      var clipped = mainChart.append("g")
        .attr("clip-path", "url(#clip)")
      var clippedBack = clipped.append("g");
      clippedFore = clipped.append("g");

      d3.json(STATIC_ENDPOINT + "/stops/" + d.gtfs_route + ".json", function(stops) {
        vis.append("g").attr("transform","translate(1024,23)").selectAll(".stop").data(stops).enter().append("text")
            .attr("class", "stop")
            .attr("text-anchor", "begin")
            .attr("y", function(d) { return stopsScale(d.index) })
            .text(function(d) { return d.name });
        drawUnanimated();
      });

      d3.json(STATIC_ENDPOINT + "/schedules/" + d.gtfs_route + ".json", function(trips) {
        var min = d3.min(trips, function(trip) {
          return d3.min(trip.stops, function(stop) { return stop.time })
        })

        var midnight = new Date();
        midnight.setHours(0,0,0,0);
        var midnightUnix = midnight.getTime() / 1000;

        // if we're beyond the first departure of the day, do it for the current day
        // otherwise, midnight of yesterday
        if ((new Date()).getTime()/1000 - midnightUnix < min) {
          midnightUnix = midnightUnix - 24 * 60 * 60;
        }

        for (var trip in trips) {
          for (var stop in trips[trip].stops) {
            trips[trip].stops[stop].time = trips[trip].stops[stop].time + midnightUnix;
          }
        }

        clippedBack.selectAll(".guide").data(trips).enter().append("path")
          .attr("class","guide")
          .classed("outbound", function(d) { return d.dir == "0" })
          .classed("inbound", function(d) { return d.dir == "1" });
        vis.selectAll(".outbound").classed("hidden",inbound);
        drawUnanimated();
      });
    }); // /selection.each
    getPastData();
  }

  // copied

  function vehicleSymbolTransform(d) {
    var x1 = timeScale(d.one.time*1000);
    var y1 = stopsScale(d.one.index);
    var x2 = timeScale(d.two.time*1000);
    var y2 = stopsScale(d.two.index);
    var angle = Math.atan2(-(y2-y1),x2-x1)*180/Math.PI;
    return "translate(" + x2 + "," + y2 + ") rotate(" + -angle + ")";
  }

  function subDraw () {
    clippedFore.selectAll(".vehiclePath").data(data, function(d) { return d.key }).enter()
      .append("path")
      .attr("id", function(d) { return "vp_" + d.key })
      .attr("class","vehiclePath")
      .classed("inbound", function(d) { return d.run.dir == 0 })
      .classed("outbound", function(d) { return d.run.dir == 1 })

    clippedFore.selectAll(".vehicleSymbol").data(ends, function(d) { return d.key }).enter()
      .append("polygon")
      .attr("class", "vehicleSymbol")
      .attr("points", "0,4 8,0 0,-4")
      .classed("inbound", function(d) { return d.dir == 0 })
      .classed("outbound", function(d) { return d.dir == 1 })

    vis.selectAll(".time.axis").call(axis);
    vis.selectAll(".vehiclePath").attr("d", function(d) { return line(d.run.states); })
    vis.selectAll(".guide").attr("d", function(d) { return line(d.stops) });
  }

  function drawUnanimated() {
    subDraw();
    vis.selectAll(".vehicleSymbol")
      .style("fill", "black")
      .attr("transform", vehicleSymbolTransform);
  }

  function drawAnimated() {
    subDraw();
    vis.selectAll(".vehicleSymbol")
      .style("fill", "#96ff30")
      .transition()
      .duration(1000)
      .style("fill", "black")
      .attr("transform", vehicleSymbolTransform);
  }

  my.d = drawUnanimated;

  function endsWithFlattenedData(flattened) {
    var ends = [];
    for (l in flattened) {
      var states = flattened[l].run.states;
      if (states.length == 1) {
        ends.push({"key":flattened[l].key,
                   "two":states[0],
                   "one":states[0],
                   "dir":flattened[l].run.dir});
      } else {
        ends.push({"key":flattened[l].key,
                   "two":states[states.length - 1],
                   "one":states[states.length - 2],
                   "dir":flattened[l].run.dir});
      }
    }
    return ends;
  }

  function getPastData() {
    d3.json(LIVE_ENDPOINT + "/locations.json?route=" + nextbus_route, function(response) {

      var flattened = [];
      for(var key in response) {
        flattened.push({"run":response[key],"key":key});
      }

      data = flattened;

      var max = d3.max(flattened, function(run) {
        return d3.max(run.run.states, function(state) { return state.time })
      })

      lastTime = (max + 60 * 15) * 1000;

      ends = endsWithFlattenedData(data);
      drawUnanimated();
      vis.selectAll(".outbound").classed("hidden",inbound);
    });
  }

  function getDataSince(timestamp) {
    d3.json(LIVE_ENDPOINT + "/locations.json?route=" + nextbus_route + "&after=" + Math.floor(timestamp/1000), function(response) {
      // delta join.
      for (var run in response) {
        var match = data.filter(function(d) { return d.key == run})
        if (match.length > 0) {
          for (var stat in response[run].states) {
            match[0].run.states.push(response[run].states[stat]);
          }
        } else {
          data.push({"run":response[run].key,"run":response[run]});
        }
      }
      ends = endsWithFlattenedData(data);
      drawAnimated();
    });
  }


  my.update = function(timestamp) {
    getDataSince(timestamp);
  }

  return my;
}


