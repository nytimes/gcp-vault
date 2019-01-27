package main

import (
	"log"

	"github.com/NYTimes/gizmo/server/kit"

	gizmoexample "github.com/NYTimes/gcp-vault/examples/gizmo-kit-example"
)

func main() {
	svc, err := gizmoexample.NewService()
	if err != nil {
		log.Fatal("unable to init service: ", err)
	}

	err = kit.Run(svc)
	if err != nil {
		log.Fatal("problems running service: ", err)
	}
}
