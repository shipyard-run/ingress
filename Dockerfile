FROM golang:latest

COPY . /go/src/github.com/shipyard-run/ingress

WORKDIR /go/src/github.com/shipyard-run/ingress

RUN CGO_ENABLED=0 go build -o bin/ingress main.go

FROM alpine:latest

RUN apk add -u socat curl

# Install Kubectl
RUN curl -sLO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl && \
  chmod +x ./kubectl && \
  mv ./kubectl /usr/local/bin

COPY --from=0 /go/src/github.com/shipyard-run/ingress/bin/ingress /usr/local/bin/ingress

ENTRYPOINT [ "ingress" ]