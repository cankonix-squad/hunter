# CI/CD operations

`CI` runs for every push to `main` and every pull request. Production deployment is deliberately
manual through the `Deploy production` workflow because release approval, the seven-day source
stability observation, classification approval, backup evidence, and UAT cannot be inferred from a
push event.

## GitHub environment configuration

Create a GitHub environment named `production` and require at least one reviewer. Configure these
environment variables:

- `DEPLOY_HOST`: `187.77.127.202`
- `DEPLOY_USER`: a dedicated least-privilege deployment account (do not use `root` long-term)
- `DEPLOY_PATH`: `/root/cankonix-node/apps/cankonix-hunter` while the current root-owned layout is
  retained; move it under the deployment account before removing root access

Configure these environment secrets:

- `DEPLOY_SSH_PRIVATE_KEY`: private half of a dedicated deployment-only SSH key
- `DEPLOY_KNOWN_HOSTS`: pinned `ssh-keyscan` output verified against the host key fingerprint through
  a separate trusted channel

Do not store a server password, production application secret, database credential, or TLS private
key in GitHub. The deployment host owns those values through the secret files documented in
`SECRETS.md`.

## Host preparation

The deployment repository must have a clean worktree and an `origin` remote that can fetch the
public repository. Docker Engine, Compose v2, the external Traefik network, and production secret
files must already exist. Create two mode-0600 configuration files outside Git tracking:

- `.env.traefik` using `.env.traefik.example` as the non-secret template
- `.env.release` containing the non-secret evidence references required by
  `scripts/release-preflight.sh`; image refs/digests and hosted CI evidence are injected by the
  workflow and must not be duplicated there

Authenticate the host to `ghcr.io` with a package-read token owned by the deployment account. Never
put that token in `.env.release` or a Compose file.

## Release procedure

1. Require the `CI` workflow to pass for the exact full commit SHA.
2. Open Actions → Deploy production → Run workflow and enter that SHA.
3. Review the production environment approval request and the preflight evidence.
4. Approve only during the agreed change window.
5. Retain the workflow URL, commit, image digests, smoke result, and operator decision as evidence.

The workflow builds four images, publishes them to GitHub Container Registry, deploys digest-pinned
references, and runs HTTPS health/readiness/security-header smoke checks. Any missing approval,
failed gate, dirty host checkout, changed SSH host key, failed pull, or unhealthy endpoint stops the
deployment.

