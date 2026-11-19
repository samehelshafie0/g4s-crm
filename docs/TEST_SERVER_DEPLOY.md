# Client test server — deployment runbook

For the client's **test (UAT) server**, not production. The same steps apply to production later, without the test banner and with business sign-off first.

Branch: `main`. The stack is the repository's `docker-compose.yml`: PostgreSQL, Redis, the Go API and nginx serving the app. Only nginx is exposed.

## 1. Server prerequisites (Ubuntu 24.04 droplet)

As `root` on a new droplet (1 vCPU / 2 GB or larger):

```bash
apt update && apt -y upgrade
apt -y install git make openssl ca-certificates curl ufw
curl -fsSL https://get.docker.com | sh
```

Add 2 GB of swap. Building the images, and LibreOffice when it converts quote attachments, need more memory than the app uses at rest:

```bash
fallocate -l 2G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

Open only SSH and the web ports:

```bash
ufw allow OpenSSH && ufw allow 80/tcp && ufw allow 443/tcp && ufw --force enable
```

Check: `docker compose version` prints a version, and `free -h` shows the swap.

## 2. Get the code and keys

```bash
git clone https://github.com/samehelshafie0/g4s-crm.git /opt/g4s-crm
cd /opt/g4s-crm
git checkout main
make setup
```

If the repository is private, GitHub asks for a username and a **personal access token** (not your password). Create a fine-grained token with read-only *Contents* access to this repository only.

`make setup` creates the JWT signing key pair in `crm-api/keys/` (git-ignored). Keep these keys; replacing them signs everyone out.

## 3. Configure `.env`

```bash
cp .env.example .env
chmod 600 .env
```

Set at least:

| Setting | Value |
|---|---|
| `APP_ENV` | `production` |
| `DB_PASSWORD` | output of `openssl rand -hex 24` |
| `JWT_REFRESH_SECRET` | output of `openssl rand -hex 32` |
| `ENVIRONMENT_LABEL` | `Test environment` — shown on every page so testers never mistake it for the live system |
| `FRONTEND_PORT` | `80`, or another port if a TLS proxy sits in front |
| `CORS_ALLOWED_ORIGINS` | leave empty |

The API **refuses to start in production** with a missing, `change-me` or short secret, a default database password, or `*` as an allowed origin. Docker Compose also stops if `DB_PASSWORD` or `JWT_REFRESH_SECRET` is unset.

## 4. Choose the data and start

Build the images (about 10–15 minutes on 1 vCPU), start the database, migrate, then start everything.

```bash
docker compose build
make key-permissions
docker compose up -d postgres redis
make migrate
```

`make key-permissions` gives the private key to the user the API container runs as. Without it the API cannot read the key on Linux and stops at start-up. Docker on macOS does not enforce this, so a Mac rehearsal will not show it.

Then choose one of:

- **Fresh test data (recommended for client UAT):**
  ```bash
  make seed-catalog
  ADMIN_EMAIL=admin@client-test.example ADMIN_PASSWORD='<12+ characters>' ADMIN_FIRST_NAME=Test ADMIN_LAST_NAME=Admin make seed
  ```
  `seed-catalog` loads the reference catalog (manufacturers, products, services, FX). `seed` creates the first administrator once.
- **A copy of the current live data:** back up live (`scripts/backup.sh` on the live machine), copy the files over, restore with `scripts/restore.sh`, then run `make migrate`. Migration 11 stops while opportunity *op2* holds two quotes (QT-2026-0002 and QT-2026-0004); delete the one the owner does not keep first. Migration 18 then numbers opportunities (OPP-2026-0001…) and renumbers each quote to share its opportunity's number.

```bash
docker compose up -d
make status
```

All four services should show `healthy` within a minute. Open `http://<server-ip>/` and sign in as the administrator; the "Test environment" label shows at the bottom of every page.

## 5. Tester accounts

Under **Users**, create one account per tester with the role they will test: sales executive, sales manager, purchasing officer, purchasing manager, warehouse manager, project manager. Share passwords privately, never in the test plan.

Testers should know:
- Five wrong passwords for one account from one address lock that account for 15 minutes. An administrator cannot unlock it early; wait or use another account.
- Password reset by email is not configured. An administrator resets passwords under Users.

## 6. Backups

```bash
./scripts/backup.sh
crontab -e   # add: 0 2 * * * cd /opt/g4s-crm && ./scripts/backup.sh >> backups/backup.log 2>&1
```

Each run writes `backups/g4s_crm-<time>.dump` (database) and `backups/storage-<time>.tgz` (uploaded documents), checks both can be read back, and keeps 14 days (`BACKUP_KEEP_DAYS`). **Copy backups off the server** regularly. A backup on the same disk does not survive the disk.

Restore (replaces the database and documents; asks you to type `RESTORE`):

```bash
./scripts/restore.sh backups/g4s_crm-<time>.dump backups/storage-<time>.tgz
make migrate
```

**Never run `make clean` on the server.** It runs `docker compose down -v`, which deletes the database and every uploaded document.

## 7. HTTPS

Testers' passwords cross the network, so switch to HTTPS as soon as the server has a domain name.

1. In your DNS, add an **A record** for, e.g., `crm-test.example.com` pointing to the droplet's IP. Wait until `ping crm-test.example.com` answers from that IP.
2. Move the app behind a local port. In `.env` set `FRONTEND_PORT=127.0.0.1:8080`, then run `docker compose up -d`.
3. Install Caddy, which obtains and renews the certificate automatically:
   ```bash
   apt -y install debian-keyring debian-archive-keyring apt-transport-https
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' > /etc/apt/sources.list.d/caddy-stable.list
   apt update && apt -y install caddy
   printf 'crm-test.example.com {\n\treverse_proxy 127.0.0.1:8080\n}\n' > /etc/caddy/Caddyfile
   systemctl reload caddy
   ```
4. Open `https://crm-test.example.com/`.

The API trusts forwarded client addresses only from private and loopback addresses, so a proxy on the same host keeps sign-in limits per real client.

## 8. Updating the test server

For the 19 November meeting-fix/price-history batch, use `feature/opportunity-flow`. Back up before updating, build while the existing app is running, then stop the web/API services while migrations run. Keep PostgreSQL, Redis, uploaded documents and signing keys in place.

```bash
set -euo pipefail
cd /opt/g4s-crm
test -z "$(git status --porcelain --untracked-files=no)" || { echo "Tracked server files have local edits; stop and review them."; exit 1; }
./scripts/backup.sh
git fetch origin
git switch feature/opportunity-flow
git pull --ff-only origin feature/opportunity-flow
docker compose build api frontend
make key-permissions
docker compose stop frontend api
make migrate
docker compose up -d --wait --wait-timeout 180
docker compose ps
docker compose exec -T api wget -qO- http://127.0.0.1:8080/ready
docker compose exec -T frontend wget -qO- http://127.0.0.1/health
```

The new batch requires migrations 19 and 20. If any command fails, stop and inspect the error before continuing. In particular, do not start the new API after a failed migration. Migrations only move forward. To roll back, restore the backup taken before the update and redeploy the previous commit.

## 9. Health checks

- `http://<server>/health`: nginx is up.
- `docker compose ps`: all services `healthy`. The API reports healthy only when every migration is applied.
- `docker compose logs --tail 100 api`: request log and startup errors, such as the unsafe-configuration refusal.
