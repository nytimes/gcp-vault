package main

import (
	"log"

	"github.com/nytimes/gizmo/server/kit"

	gizmoexample "github.com/nytimes/gcp-vault/examples/gizmo-kit-example"
)

func main() {
	svc, err := gizmoexample.NewService()
	if err != nil {
		log.Fatal("unable to init service: ", err)
	}

	err = kit.Run(svc)
	if err != nil {
		log.Fatal("prboblems running service: ", err)
	}
}
