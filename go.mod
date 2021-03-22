module github.com/shipyard-run/ingress

go 1.13

require (
	github.com/apex/log v1.9.0 // indirect
	github.com/hashicorp/go-hclog v0.15.0
	github.com/prometheus/common v0.10.0
	github.com/shipyard-run/shipyard v0.3.2-0.20210322075841-a216c25de911
	github.com/spf13/pflag v1.0.5
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)

//replace github.com/shipyard-run/shipyard => ../shipyard
