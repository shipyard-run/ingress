build_docker:
	docker build -t registry.shipyard.run/ingress:latest .
	docker tag registry.shipyard.run/ingress:latest gcr.io/shipyard-287511/ingress:latest

push_docker:
	docker push gcr.io/shipyard-287511/ingress:latest

build_and_push_docker: build_docker push_docker
