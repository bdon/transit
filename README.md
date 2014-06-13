# bdon.org/transit #

## Prerequisites ##
* a Go distribution for Linux or Mac
* the `gox` cross-compiler: `go get gox` 
* Ansible 1.6: `brew install ansible` or via the PPA
* An Ubuntu server with at least 512 MB of memory and 20GB disk

## How to use ##
`go build .` will create the `transit` binary.

`./transit --emitFiles`: will emit schedules, stops, and route JSON files into the `static/` directory.

`./transit` : will poll NextBus, write history to `static/history`, and serve location requests on port 8080.

## Run the tests ##

API tests: `go test .`

Client tests: open `www/test/index.html` in your browser.