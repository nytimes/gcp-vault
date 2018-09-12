package main

import (
	"log"

	"github.com/NYTimes/gizmo/server/kit"
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
