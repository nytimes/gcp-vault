The app.yaml here is meant for actual deployments.

The `make_local_yaml` script here is a helper for devs to inject their github PAT into the local App Engine environment when running the service locally.

Users should first run `make_local_yaml` to generate a local.yaml file and then run the service via `dev_appserver.py local.yaml`
