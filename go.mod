module github.com/shipyard-run/ingress

go 1.13

require (
	github.com/hashicorp/go-hclog v0.14.1
	github.com/prometheus/common v0.7.0
	github.com/shipyard-run/shipyard v0.1.16-0.20201105185416-b6453f08f9c0
	github.com/spf13/pflag v1.0.5
)

//replace github.com/shipyard-run/shipyard => ../shipyard
