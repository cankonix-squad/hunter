# Production secrets

Production processes accept sensitive values only from files mounted through the
`*_FILE` variables. This keeps values out of Compose interpolation, container environment,
process listings, and routine configuration output. The application rejects ambiguous
configuration when both the direct variable and its `_FILE` counterpart are set.

## Required files

Create these files outside the repository with mode `0600` and ownership limited to the
deployment operator:

- PostgreSQL password: at least 32 random bytes;
- complete `DATABASE_URL`, using the same PostgreSQL password;
- JWT signing secret: at least 32 random bytes, base64 encoded (43 or more characters).

Generate independent values with a cryptographically secure generator, for example
`openssl rand -base64 48`. Never paste values into `.env`, Compose YAML, tickets, logs,
or this repository. Point only the following host variables at their file paths:

```text
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE=/secure/path/postgres_password
CANKONIX_DATABASE_URL_SECRET_FILE=/secure/path/database_url
CANKONIX_JWT_SECRET_FILE=/secure/path/jwt_secret
```

Then render and deploy both Compose files:

```sh
docker compose -f compose.yaml -f compose.production.yaml config
docker compose -f compose.yaml -f compose.production.yaml up -d
```

The optional bootstrap administrator must normally remain disabled in production. For the
first controlled bootstrap only, set `BOOTSTRAP_ADMIN_EMAIL` and point
`CANKONIX_BOOTSTRAP_ADMIN_PASSWORD_SECRET_FILE` at a random password of at least 20
characters. Remove both settings and the file after the account is created. SMTP and webhook
credentials use `CANKONIX_SMTP_PASSWORD_SECRET_FILE` and
`CANKONIX_WEBHOOK_SECRET_FILE` when their non-dry-run adapters are enabled.

## Rotation

1. Create the replacement secret as a new version in the deployment secret store.
2. For database credentials, change the database role password and `DATABASE_URL` in one
   maintenance window, then restart all database clients. Verify `/readyz` before removing
   the old version.
3. For JWT, replace the file and restart API, collector, and notifier. Existing access tokens
   become invalid; refresh sessions can obtain tokens signed by the replacement secret.
4. For SMTP or webhook, replace the file and restart notifier, then verify one controlled
   delivery without recording the credential.
5. Record operator, time, secret version identifier, service health, and rollback result—never
   the value. Immediately rotate any credential suspected of disclosure.

Secret files must come from the production platform secret manager or an encrypted deployment
store. The Compose file provides the injection boundary; it is not itself a secret store.
