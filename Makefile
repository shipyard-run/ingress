build_docker:
	docker build -t docker.pkg.github.com/shipyard-run/ingress/ingress:latest .

push_docker:
	docker push docker.pkg.github.com/shipyard-run/ingress/ingress:latest