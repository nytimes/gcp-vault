# The basic App Engine Standard Environment Example

This example shows how to use basic HTTP middleware to fetch secrets from Vault.

To run this service, you must be using [Google Cloud SDK](https://cloud.google.com/appengine/docs/standard/go/download) >= `162.0.0` or the "original" App Engine Go SDK >= `1.9.56`.

To run this against your own Vault installation, update the values in `app.yaml` for deployment and `make_local_yaml.sh` for local development.

The `run_local.sh` script wraps both the Vault login and App Engine's `dev_appserver.py` command to simplify running the server locally.
