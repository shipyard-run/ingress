FROM golang:latest as build

COPY . /go/src/github.com/shipyard-run/ingress

WORKDIR /go/src/github.com/shipyard-run/ingress

RUN CGO_ENABLED=0 go build -o bin/ingress main.go

FROM alpine:latest as base

RUN apk add -u socat curl

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG BUILDPLATFORM
ARG BUILDARCH

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM $TARGETARCH $TARGETVARIANT"  

COPY --from=build /go/src/github.com/shipyard-run/ingress/bin/ingress /usr/local/bin/ingress

# Install Kubectl
RUN curl -sLO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/$TARGETPLATFORM/kubectl && \
  chmod +x ./kubectl && \
 
mv ./kubectl /usr/local/bin

ENTRYPOINT [ "ingress" ]