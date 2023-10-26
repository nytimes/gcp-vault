module github.com/NYTimes/gcp-vault/examples

go 1.12

replace github.com/NYTimes/gcp-vault => ../

require (
	github.com/NYTimes/gcp-vault v0.0.0-00010101000000-000000000000
	github.com/NYTimes/gizmo v0.4.3
	github.com/NYTimes/marvin v0.2.1
	github.com/go-kit/kit v0.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	google.golang.org/appengine v1.6.8
	google.golang.org/grpc v1.58.3
)
