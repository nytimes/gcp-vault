// Package gcpvault provides tools for securely retrieving secrets from Vault while
// running on Google Cloud infrastructure.
//
// To use this library, users must follow the instructions for enabling GCP
// authentication: https://www.vaultproject.io/docs/auth/gcp.html
//
// For local development, users should use something like Github personal access tokens
// to log into vault before injecting their Vault login token into the local environment.
package gcpvault // import "github.com/nytimes/gcp-vault"
