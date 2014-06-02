all:
	gox -output="bin/transit_timelines" -osarch="linux/amd64"
	ansible-playbook -v -i hosts contrib/playbook.yml


