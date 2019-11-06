module github.com/NYTimes/gcp-vault/examples

go 1.13

replace github.com/NYTimes/gcp-vault => ../

require (
	cloud.google.com/go/logging v1.0.0 // indirect
	github.com/NYTimes/gcp-vault v0.2.2
	github.com/NYTimes/gizmo v0.4.3
	github.com/NYTimes/marvin v0.2.1
	github.com/go-kit/kit v0.8.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/pkg/errors v0.8.1
	google.golang.org/appengine v1.6.1
	google.golang.org/grpc v1.22.0
)
