version=v0.3.0
repo=shipyardrun/ingress

build_docker:
	docker build -t ${repo}:${version} .

push_docker:
	docker push ${repo}:${version}

build_and_push_docker: build_docker push_docker
