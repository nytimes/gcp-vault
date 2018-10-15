# gcpvault

[![GoDoc](https://godoc.org/github.com/NYTimes/gcp-vault?status.svg)](https://godoc.org/github.com/NYTimes/gcp-vault) [![Build Status](https://travis-ci.org/NYTimes/gcp-vault.svg?branch=master)](https://travis-ci.org/NYTimes/gcp-vault)

`gcpvault` provides a function for securely retrieving secrets from [HashiCorp Vault](https://www.vaultproject.io/) while running on the [Google Cloud Platform](https://cloud.google.com/) or in your local development environment.

To use this library, users must follow the Vault instructions for enabling GCP authentication: [https://www.vaultproject.io/docs/auth/gcp.html](https://www.vaultproject.io/docs/auth/gcp.html).

Under the hood, when deployed to Google Cloud this tool will use the [default application credentials](https://cloud.google.com/docs/authentication/production) to login to Vault and retrieve the specified secrets.

Since the login to Vault can be a heavy and relatively slow operation, we recommend users of the legacy [Google App Engine Standard Environment](https://cloud.google.com/appengine/docs/standard/) (Go <=1.9) call this library during [start up requests for manual scaling systems](https://cloud.google.com/appengine/docs/standard/go/how-instances-are-managed#startup) or in [warm up requests for users of automatic scaling](https://cloud.google.com/appengine/docs/standard/go/how-instances-are-managed#warmup_requests) to prevent exposing public traffic to such latencies.

## Local Development

For local development, users should use a Github personal access tokens or some similar method to [login to Vault](https://www.vaultproject.io/docs/commands/login.html) before injecting their Vault login token into the local environment.

## Unit Testing

For mocking out the services required for interacting with Vault, a [gcpvaulttest](https://godoc.org/github.com/NYTimes/gcp-vault/gcpvaulttest) package has been included to provide `httptest.Server`s for each dependency.

## Examples

Check out the [examples](https://github.com/NYTimes/gcp-vault/tree/master/examples/) directory for examples on how to use this package.
