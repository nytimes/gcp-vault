# gcpvault

[![GoDoc](https://godoc.org/github.com/NYTimes/gcp-vault?status.svg)](https://godoc.org/github.com/NYTimes/gcp-vault) [![Build Status](https://cloud.drone.io/api/badges/nytimes/gcp-vault/status.svg)](https://cloud.drone.io/nytimes/gcp-vault)

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

#Vault Token Caching

The library has an option to enable Vault Token Caching. Currently, Redis or GCS is supported for token storage. To enable token caching,
one the following environment variables should set:

**TOKEN_CACHE_STORAGE_REDIS** - Host and port for Redis '10.200.30.4:6379'

**TOKEN_CACHE_STORAGE_GCS**  - GCS bucket location where token can be stored for caching purposes. Care should be taken to make sure bucket permissions are set such that vault token is not leaked to the world.

Additional optional environment variables that control cache.

**TOKEN_CACHE_REFRESH_THRESHOLD** - How long before the token expiration should it be regenerated (in seconds). Default is 300 seconds.

**TOKEN_CACHE_KEY_NAME** - The object name to store. Default value is _token-cache_.

**TOKEN_CACHE_CTX_TIMEOUT** - This value is in seconds. Default value is 30 seconds.

**TOKEN_CACHE_STORAGE_REDIS_DB** - Database for Redis. Default is 0.

**TOKEN_CACHE_REFRESH_RANDOM_OFFSET** - Random refresh offset in seconds to avoid all the instances refreshing at once. Default is 1/2 the duration in seconds of the _TOKEN_CACHE_REFRESH_THRESHOLD_.
