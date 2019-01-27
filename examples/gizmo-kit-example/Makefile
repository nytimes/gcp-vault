
build:
	@cd ./cmd/server; \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

deps:
	@export GO111MODULE=on; \
	go get -v .; \
	go mod vendor; \

proto:
	@go get github.com/NYTimes/openapi2proto/cmd/openapi2proto
	@openapi2proto -spec ./leaderboards.yaml -out ./leaderboards.proto
	@protoc --go_out=. ./leaderboards.proto

deploy-local: build
	@say -v Samantha "compiled\!"
	@cd ./cmd/server; \
	docker build  --tag gcr.io/nyt-games-dev/leaderboards:local .; say -v Samantha "image built\!";
	@docker push gcr.io/nyt-games-dev/leaderboards:local; say -v Samantha "image pushed\!";
	@cd ./cmd/server; \
	gcloud app deploy ./dev.yaml --version "local-`date +'%s'`"  --image-url gcr.io/nyt-games-dev/leaderboards:local --project nyt-games-dev --quiet; say -v Samantha "deployed\!";

test-drone: vet
	@mkdir -p /opt/nyt/etc
	@echo $(INT_TEST_CREDENTIALS) | base64 -d > /opt/nyt/etc/games-puzzles-sandbox.json
	@export GOOGLE_CLOUD_PROJECT="games-puzzles-sandbox"; \
	export GOOGLE_APPLICATION_CREDENTIALS="/opt/nyt/etc/games-puzzles-sandbox.json"; \
	export IAM_SERVICE_ACCOUNT_EMAIL="leaderboard-integration-tests@games-puzzles-sandbox.iam.gserviceaccount.com"; \
	export IAM_PROJECT="games-puzzles-sandbox"; \
	export IAM_AUDIENCE="localhost"; \
	export ID_AUDIENCE="localhost"; \
	export INVITE_URL="https://www.stg.nytimes.com/puzzles/leaderboard/invite/"; \
	export VAULT_GCP_IAM_ROLE="games-puzzles-sandbox-games-leaderboards-iam"; \
	export VAULT_ADDR=$$VAULT_ADDR; \
	export VAULT_SECRET_PATH="nytm/games-leaderboards/secret/games-puzzles-sandbox"; \
	export HTTP_PORT=8888; \
	go test -v -race -cover .

vet:
	@go vet .
	@go vet ./cmd/server

test: _vault _gcloud-login
	@vault read -format json nytm/games-leaderboards/secret/games-puzzles-sandbox | jq -r .data.IntegrationCreds | base64 -D -o $(TEST_GCP_CREDS)
	@export GOOGLE_APPLICATION_CREDENTIALS=$(TEST_GCP_CREDS); \
	export GOOGLE_CLOUD_PROJECT="games-puzzles-sandbox"; \
	export VAULT_ADDR=$$VAULT_ADDR; \
	export VAULT_LOCAL_TOKEN="`cat ~/.vault-token`"; \
	export IAM_SERVICE_ACCOUNT_EMAIL="leaderboard-integration-tests@games-puzzles-sandbox.iam.gserviceaccount.com"; \
	export IAM_PROJECT="games-puzzles-sandbox"; \
	export IAM_AUDIENCE="localhost"; \
	export ID_AUDIENCE="localhost"; \
	export VAULT_SECRET_PATH="nytm/games-leaderboards/secret/games-puzzles-sandbox"; \
	export INVITE_URL="https://www.stg.nytimes.com/puzzles/leaderboard/invite/"; \
	export HTTP_PORT=8800; \
	go test -v -cover 
	@rm $(TEST_GCP_CREDS);

build-local:
	@cd ./cmd/server; \
	go build .

run: build-local _vault _gcloud-login
	@export VAULT_ADDR=$$VAULT_ADDR; \
	export VAULT_LOCAL_TOKEN="`cat ~/.vault-token`"; \
	export GOOGLE_CLOUD_PROJECT="nyt-games-dev"; \
	export IAM_SERVICE_ACCOUNT_EMAIL="nyt-games-dev@appspot.gserviceaccount.com"; \
	export IAM_AUDIENCE="localhost"; \
	export ID_AUDIENCE="localhost"; \
	export VAULT_SECRET_PATH="nytm/games-leaderboards/secret/nyt-games-dev"; \
	export INVITE_URL="https://www.nytimes.com/puzzles/leaderboard/invite/"; \
	export NO_AUTH="true"; \
	export HTTP_PORT=8084; \
	./cmd/server/server

# Runs a given "migration" file name against the given "project"'s reading-list Spanner instance.
spanner-run-migration: _gcloud-login
	@gcloud spanner databases ddl update reading-list --instance reading-list --ddl="`cat ./ddl/$(migration).sql`" --project $(project)

_gcloud-login:
	@if [ ! -f "$(GCP_CREDS)" ]; then gcloud auth application-default login; fi

_vault:
	@export VAULT_ADDR=$$VAULT_ADDR; \
	vault login -method=github token=`cat ~/.config/vault/github` > /dev/null 2>&1;

VAULT_ADDR=https://vault-prd.dv.nyt.net
GCP_CREDS = $(HOME)/.config/gcloud/application_default_credentials.json
TEST_GCP_CREDS = $(HOME)/.config/gcloud/test_application_default_credentials.json


.PHONY: test, build, run, test-drone, _vault, gcloud
.EXPORT_ALL_VARIABLES:
