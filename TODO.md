* how do i find the route name etc?, PM, PH, C cable cars
* place to list all available routes.
* agency ids and nextbus agencys dont match

* fetch timeout
* error logging
* combine schedules and stops
* configure CORS
* idempotent line simplification - cut out points that are less than 1 minute apart? or distance?

* flesh out README
* add another agency (timezone!)

Frontend:
* show hash state in URL
* store stop_id -> latlong to show on map?

* add reset.css
* how do we define "inbound" - shape order?
* Indexes should be culled to last scheduled stop.

* fix arrows coming from nowhere
* only show arrows for currently live routes
* only show one of inbound/outbound at a time, to reduce DOM elements

* Use correct GTFS schedule.
