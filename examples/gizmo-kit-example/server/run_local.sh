#!/bin/sh

export VAULT_ADDR="https://vault.your-company.com";

vault login -method=github token=`cat ~/.config/vault/github` > /dev/null 2>&1;

export VAULT_LOCAL_TOKEN="`cat ~/.vault-token`"
export VAULT_ADDR="https://vault.your-company.com"
export VAULT_SECRET_PATH="repo-name/secret/my-secrets"

go run main.go
