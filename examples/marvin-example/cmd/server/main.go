package main

import (
	"github.com/NYTimes/marvin"

	"google.golang.org/appengine"

	marvinexample "github.com/nytimes/gcp-vault/examples/marvin-example"
)

func main() {
	marvin.Init(marvinexample.NewService())
	appengine.Main()
}
