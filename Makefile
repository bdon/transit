converge/staging:
	cd chef && ./deploy.sh bdon@transitstaging

ssh/staging:
	ssh bdon@transitstaging

sync/staging:
	rsync -avz . bdon@transitstaging:/home/bdon/go/src/github.com/bdon/jklmnt --exclude '.git'
	rsync -avz muni_gtfs bdon@transitstaging:/var/serve/

converge/prod:
	cd chef && ./deploy.sh bdon@transit

ssh/prod:
	ssh bdon@transit

sync/prod:
	rsync -avz . bdon@transit:/home/bdon/go/src/github.com/bdon/jklmnt --exclude '.git'
	rsync -avz muni_gtfs bdon@transit:/var/serve/
