var m = [16, 16, 16, 16],
    w = 1200 - m[1] - m[3],
    h = 200 - m[0] - m[2];

var svg = d3.select("#chart").append("svg:svg")
    .attr("class", "muni")
    .attr("width", w + m[1] + m[3])
    .attr("height", h + m[0] + m[2])
var vis = svg.append("svg:g")
    .attr("transform", "translate(" + m[3] + "," + m[0] + ")");

var lastTime;
var timeScale = d3.time.scale().range([0,1020]);
var stopsScale = d3.scale.linear().domain([0,1000]).range([5,h-20]);
var axis = d3.svg.axis().scale(timeScale).orient("top")
var toggles = d3.select("#toggles");

var gtfs_route_id = "1093";
var nextbus_route = "N";
var static_endpoint = "http://localhost:8081/static";
var live_endpoint = "http://localhost:8080";

var data = [];
var ends = [];
var guides = [];

var line = d3.svg.line()
  .x(function(d) { return timeScale(d.time*1000) })
  .y(function(d) { return stopsScale(d.index) })
  .interpolate("linear");

toggles.append("input")
  .attr("id","inbound")
  .attr("type", "checkbox")
  .attr("checked",true)
  .on("click", function() {
    d3.selectAll(".inbound").classed("hidden",!this.checked);
  });

toggles.append("label")
  .attr("for", "inbound")
  .text("Inbound");

toggles.append("input")
  .attr("id","outbound")
  .attr("type", "checkbox")
  .on("click", function() {
    d3.selectAll(".outbound").classed("hidden",!this.checked);
  });

toggles.append("label")
  .attr("for", "outbound")
  .text("Outbound");

function vehicleSymbolTransform(d) {
  var x1 = timeScale(d.one.time*1000);
  var y1 = stopsScale(d.one.index);
  var x2 = timeScale(d.two.time*1000);
  var y2 = stopsScale(d.two.index);
  var angle = Math.atan2(-(y2-y1),x2-x1)*180/Math.PI;
  return "translate(" + x2 + "," + y2 + ") rotate(" + -angle + ")";
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

var zoom = d3.behavior.zoom()
    .scaleExtent([0.1,0.5])
    .on("zoom", drawUnanimated);

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
var clippedMid = clipped.append("g");
var clippedFore = clipped.append("g");
var tooltip = clippedFore.append("text")
    .attr("class", "vehicleTooltip")

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
    .attr("points", "0,5 10,0 0,-5")
    .classed("inbound", function(d) { return d.dir == 0 })
    .classed("outbound", function(d) { return d.dir == 1 })

  vis.selectAll(".time.axis").call(axis);

  vis.selectAll(".vehiclePath").attr("d", function(d) { return line(d.run.states); })
  vis.selectAll(".guide").attr("d", function(d) { return line(d.stops) });
}

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
  d3.json(live_endpoint + "/locations.json?route=" + nextbus_route, function(response) {

    var flattened = [];
    for(var key in response) {
      flattened.push({"run":response[key],"key":key});
    }

    data = flattened;

    var max = d3.max(flattened, function(run) {
      return d3.max(run.run.states, function(state) { return state.time })
    })

    lastTime = (max + 60 * 15) * 1000;
    timeScale.domain([(lastTime/1000 - 60 * 60) * 1000, lastTime]);
    zoom.x(timeScale);
    zoom.scale(0.15);
    zoom.translate([700,0]);

    ends = endsWithFlattenedData(data);
    drawUnanimated();
    d3.selectAll(".outbound").classed("hidden",!this.checked);
  });
}

function getDataSince(timestamp) {
  d3.json(live_endpoint + "/locations.json?route=" + nextbus_route + "&after=" + Math.floor(timestamp/1000), function(response) {
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

getPastData();
var timestamp = new Date().getTime();
var timeTilUpdate = 9;

setInterval(function() {
  if (timeTilUpdate <= 0) {
    getDataSince(timestamp);
    timestamp = new Date().getTime();
    timeTilUpdate = 10;
  }
  timeTilUpdate--;
}, 1 * 1000)

d3.json(static_endpoint + "/schedules/" + gtfs_route_id + ".json", function(trips) {
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
  d3.selectAll(".outbound").classed("hidden",!this.checked);
});

d3.json(static_endpoint + "/stops/" + gtfs_route_id + ".json", function(stops) {
  vis.append("g").attr("transform","translate(1024,23)").selectAll(".stop").data(stops).enter().append("text")
      .attr("class", "stop")
      .attr("text-anchor", "begin")
      .attr("y", function(d) { return stopsScale(d.index) })
      .text(function(d) { return d.name });
  drawUnanimated();
});

