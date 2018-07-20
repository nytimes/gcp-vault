# gaevault

[![GoDoc](https://godoc.org/github.com/NYTimes/gaevault?status.svg)](https://godoc.org/github.com/NYTimes/gaevault)

`gaevault` provides a function for securely retrieving secrets from [HashiCorp Vault](https://www.vaultproject.io/) while running in the [Google App Engine Standard Environment](https://cloud.google.com/appengine/docs/standard/) or in your local development environment.

To use this library, users must follow the Vault instructions for enabling GCP authentication: [https://www.vaultproject.io/docs/auth/gcp.html](https://www.vaultproject.io/docs/auth/gcp.html).

Under the hood, when deployed to the GAE standard environment this tool will use the GCP project's default App Engine service account ({your-project-name}@appspot.gserviceaccount.com) to sign a JWT and log into Vault.

Since the login to Vault can be a heavy and relatively slow operation, we recommend users call this library during [start up requests for manual scaling systems](https://cloud.google.com/appengine/docs/standard/go/how-instances-are-managed#startup) or in [warm up requests for users of automatic scaling](https://cloud.google.com/appengine/docs/standard/go/how-instances-are-managed#warmup_requests) to prevent exposing public traffic to such latencies.

For local development, users should use a Github personal access tokens or some similar method to [login to Vault](https://www.vaultproject.io/docs/commands/login.html) before injecting their Vault login token into the local environment.

## Examples

Check out the [examples](https://github.com/NYTimes/gae-vault/tree/master/examples/) directory for examples on how to use this package.
