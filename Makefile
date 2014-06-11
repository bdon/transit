all:
	gox -output="bin/transit_timelines" -osarch="linux/amd64"
	ansible-playbook -i hosts contrib/playbook.yml

cp:
	cp www/index.html ~/workspace/bdon.org/content/transit/
	cp www/src/index.js ~/workspace/bdon.org/content/transit/src/index.js
	cp www/style.css ~/workspace/bdon.org/content/transit/style.css
