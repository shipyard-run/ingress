version=v0.3.0
repo=shipyardrun/ingress

build_docker:
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx create --name multi || true
	docker buildx use multi
	docker buildx inspect --bootstrap
	docker buildx build --platform linux/arm64,linux/amd64 \
		-t ${repo}:${version} \
    -f ./Dockerfile \
		--push \
    .
	docker buildx rm multi