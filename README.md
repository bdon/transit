The 'ttimeline' binary has 3 modes:

Modes

Required Arguments
--gtfs DIR : directory containing uncompressed GTFS files

Optional:
--static DIR : the static directory to serve from. 
  the program stores its state here on quit.
  intended to serve this through Nginx.

if not provided, will not read initial state/persist its state to disk.


'ttimeline' - runs in "Collect" mode - no webserver.
'ttimeline -transform'
'ttimeline -serve 8080'
