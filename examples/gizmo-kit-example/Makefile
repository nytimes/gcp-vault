# Cross compiles the server for Linux, builds a "local" docker image, pushes it and deploys it to App Engine
deploy-local: build _gcloud-login
	@say -v Samantha "compiled\!"
	@cd cmd/server; \
	docker build  --tag $$LOCAL_IMAGE .; say -v Samantha "image built\!";
	@docker push $$LOCAL_IMAGE; say -v Samantha "image pushed\!";
	cd cmd/server; \
	gcloud app deploy ./app.yaml --version "local-`date +'%s'`"  --image-url $$LOCAL_IMAGE --project my-gcp-project --quiet; say -v Samantha "deployed\!";
	@rm cmd/server/server;

build:
	@cd ./cmd/server; \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

build-local:
	@cd ./cmd/server; \
    go build .

run: _vault build-local
	@export VAULT_ADDR=$$VAULT_ADDR; \
	export VAULT_LOCAL_TOKEN="`cat ~/.vault-token`"; \
	export VAULT_SECRET_PATH="repo-name/secret/my-secrets"; \
	cd ./cmd/server; \
	./server;

test:
	@go test ./... -v -cover

_vault:
	@export VAULT_ADDR=$$VAULT_ADDR; \
	vault login -method=github token=`cat ~/.config/vault/github` > /dev/null 2>&1;

_gcloud-login:
	@if [ ! -f "$(HOME)$(GCP_CREDS)" ]; then gcloud auth application-default login; fi

VAULT_ADDR=https://vault.example.com
GCP_CREDS=/.config/gcloud/application_default_credentials.json
LOCAL_IMAGE=gcr.io/my-gcp-project/gcpvault-example:local

.EXPORT_ALL_VARIABLES:
