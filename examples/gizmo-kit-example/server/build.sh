#!/bin/sh

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

docker build  --tag gcr.io/$1/gizmoexample .

gcloud docker push gcr.io/$1/gizmoexample
