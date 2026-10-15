# Client test server — deployment runbook

For the client's **test (UAT) server**, not production. The same steps apply to production later, without the test banner and with business sign-off first.

Branch: `feature/opportunity-flow`. The stack is the repository's `docker-compose.yml`: PostgreSQL, Redis, the Go API and nginx serving the app. Only nginx is exposed.

## 1. Server prerequisites

- Linux host with Docker Engine and the Compose plugin, `git` and `openssl`.
- Port 80 open to the client's testers. For HTTPS see section 7.
- About 2 GB RAM and 10 GB disk to start; uploads and backups grow over time.

## 2. Get the code and keys

```bash
git clone https://github.com/samehelshafie0/g4s-crm.git /opt/g4s-crm
cd /opt/g4s-crm
git checkout feature/opportunity-flow
make setup
```

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

Start the database, migrate, then start everything.

```bash
docker compose up -d postgres redis
make migrate
```

Then choose one of:

- **Fresh test data (recommended for client UAT):**
  ```bash
  make seed-catalog
  ADMIN_EMAIL=admin@client-test.example ADMIN_PASSWORD='<12+ characters>' ADMIN_FIRST_NAME=Test ADMIN_LAST_NAME=Admin make seed
  ```
  `seed-catalog` loads the reference catalog (manufacturers, products, services, FX). `seed` creates the first administrator once.
- **A copy of the current live data:** back up live (`scripts/backup.sh` on the live machine), copy the files over, restore with `scripts/restore.sh`, then run `make migrate`. Migration 11 stops while opportunity *op2* holds two quotes (QT-2026-0002 and QT-2026-0004); delete the one the owner does not keep first.

```bash
docker compose up -d --build
make status
```

Open `http://<server>/` and sign in as the administrator.

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

Testers' passwords cross the network, so use HTTPS as soon as the server has a domain name. Run a TLS proxy on the host, for example Caddy (`reverse_proxy 127.0.0.1:8080` with automatic certificates), and set `FRONTEND_PORT=127.0.0.1:8080` in `.env`. The API trusts forwarded client addresses only from private and loopback addresses, so a proxy on the same host or private network keeps sign-in limits per real client.

## 8. Updating the test server

```bash
./scripts/backup.sh
git pull
docker compose up -d --build
make migrate
docker compose restart api
```

Migrations only move forward. To roll back, restore the backup taken before the update and redeploy the previous commit.

## 9. Health checks

- `http://<server>/health`: nginx is up.
- `docker compose ps`: all services `healthy`. The API reports healthy only when every migration is applied.
- `docker compose logs --tail 100 api`: request log and startup errors, such as the unsafe-configuration refusal.
