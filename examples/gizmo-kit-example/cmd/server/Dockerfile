FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD server /server

ENV VAULT_ADDR "https://vault.your-company.com"
ENV VAULT_SECRET_PATH "repo-name/secret/my-secrets" 
ENV VAULT_GCP_IAM_ROLE "my-svc-acct-vault-role"

ENTRYPOINT["/server"]
