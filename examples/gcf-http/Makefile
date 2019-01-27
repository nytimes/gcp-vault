
test: _vault
	@export VAULT_ADDR=$$VAULT_ADDR; \
	export VAULT_LOCAL_TOKEN="`cat ~/.vault-token`"; \
	export VAULT_SECRET_PATH="repo-name/secret/my-secrets"; \
	go test ./... -v -cover 

_vault:
	@export VAULT_ADDR=$$VAULT_ADDR; \
	vault login -method=github token=`cat ~/.config/vault/github` > /dev/null 2>&1;

VAULT_ADDR=https://vault.example.com

.EXPORT_ALL_VARIABLES:
