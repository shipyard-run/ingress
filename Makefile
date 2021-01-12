build_docker:
	docker build -t shipyardrun/ingress:latest .

push_docker:
	docker push shipyardrun/ingress:latest

build_and_push_docker: build_docker push_docker
