module github.com/NYTimes/gcp-vault/examples

go 1.13

replace github.com/NYTimes/gcp-vault => ../

require (
	cloud.google.com/go/logging v1.0.0 // indirect
	github.com/NYTimes/gcp-vault v0.3.3
	github.com/NYTimes/gizmo v0.4.3
	github.com/go-kit/kit v0.8.0
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191105034135-c7e5f84aec59 // indirect
	golang.org/x/net v0.0.0-20191105084925-a882066a44e0 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/api v0.13.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.25.0
	gopkg.in/square/go-jose.v2 v2.4.0 // indirect
)
