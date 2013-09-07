converge/staging:
	cd chef && ./deploy.sh bdon@transitstaging

ssh/staging:
	ssh bdon@transitstaging

converge/prod:
	cd chef && ./deploy.sh bdon@transit

ssh/prod:
	ssh bdon@transit
