# TableKit

Talk to your database from ChatGPT and Claude. TableKit is a small,
self-hostable MCP server you point at your Postgres, MySQL and MariaDB databases
— then ask questions in plain English and get back tables and interactive charts.
Every query it runs is read-only, so the assistant can look but never touch.

You run it on your own infrastructure. Your connection strings and data stay
with you; the only thing the assistant ever sees is the rows your read-only
queries return.

## What it does

- Reads your schema so the assistant knows your tables, columns and relationships.
- Turns questions into read-only SQL — `SELECT`s only, no writes, no DDL.
- Renders results as interactive charts right inside the chat, using MCP Apps —
  switch between Chart, Table and SQL views, hover for details, and export the
  full result to CSV or JSON without leaving the conversation.
- Holds several named database connections at once, so you can keep `analytics`
  and `billing` side by side and tell the assistant which one to use.
- Works with ChatGPT on the web and Claude on web and desktop.

<!-- placeholder: replace with a real screenshot once captured -->
![A chart rendered from a query inside the chat](docs/chart.png)

## Quickstart

You'll need Docker. Declare your databases in a YAML file, give TableKit the URL
it will be reachable on, and bring it up.

Databases live in `databases.yaml` (by default `./databases.yaml` in the working
directory; point elsewhere with `DATABASES_FILE`). Each entry is keyed by the name the assistant
will use:

```yaml
# databases.yaml
databases:
  analytics:
    type: postgres
    details:
      host: db-host
      database: analytics
      username: readonly
      password: { from: env, env: ANALYTICS_PW }   # or a literal string
  billing:
    type: mysql        # or: mariadb
    connectionString: mysql://readonly:pw@db-host:3306/billing
```

```bash
# .env
PUBLIC_BASE_URL=https://tablekit.your-host.com
ANALYTICS_PW=...        # any secrets your databases.yaml reads via { from: env }
```

```bash
docker compose up --build
```

See [Databases](#databases) below for the full set of fields (TLS, SSH tunnels,
secret sources).

Check it's alive:

```bash
curl localhost:8081/health
# {"status":"OK", ...}
```

A couple of notes worth knowing up front: TableKit only ever issues read-only
SQL, but it's still good practice to hand it a database user that's read-only too
— defense in depth. And while it runs fine on `localhost` for a try-out, you'll
want it behind HTTPS before connecting a real assistant to it.

## Connecting ChatGPT or Claude

In ChatGPT, add a custom connector; in Claude, add a custom MCP server. Either
way, point it at:

```
https://tablekit.your-host.com/mcp
```

The first time a client connects, it walks through the OAuth flow on its own —
there's nothing to copy-paste. The first client to finish authorizing becomes
the paired client. That pairing is deliberate: a second client trying to connect
gets an "already paired" page instead of access, so a stray link can't quietly
attach itself to your database. If you're connecting both ChatGPT and Claude, or
re-pairing after a reinstall, open up pairing again with the CLI below.

## Configuration

Everything is set through the environment.

| Variable          | Default                    | What it's for                                   |
|-------------------|----------------------------|-------------------------------------------------|
| `DATABASES_FILE`  | `./databases.yaml`  | The YAML file declaring your databases (see below), relative to the working directory. Resolved by base name, so `.yaml` or `.yml` both work; having both at once is a fatal startup error. A missing file just means no databases. |
| `PUBLIC_BASE_URL` | `http://localhost:8080`    | The URL clients reach TableKit on.              |
| `APP_PORT`        | `8080`                     | MCP + OAuth listener.                           |
| `CONTROL_PORT`    | `8081`                     | Health and ops listener.                        |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | — / `5432` / — / — / `tablekit` / `disable` | Structured connection to TableKit's own state database (Postgres). When `DB_HOST`, `DB_USER` and `DB_PASSWORD` are all set, they take precedence over `DATABASE_URL`. |
| `DATABASE_URL`    | `postgres://postgres@localhost:5432/tablekit?sslmode=disable` | Full `postgres://` DSN for TableKit's own state database, used when the structured `DB_*` variables above are not all set. |
| `SIGNING_KEY`     | **required**               | Base64 HS256 signing key (e.g. `openssl rand -base64 32`). TableKit does not generate one — startup fails if it is unset. Share the same value across instances. Short keys are zero-padded to 32 bytes. |
| `ACCESS_TTL`      | `15m`                      | Access token lifetime.                          |
| `REFRESH_TTL`     | `168h`                     | Refresh token lifetime.                         |

### Databases

Databases are declared in `databases.yaml` as a map keyed by the name the
assistant uses with `run_query` / `list_databases`. Each entry sets a `type`
(`postgres`, `mysql`, or `mariadb`) and connects either with structured
`details` **or** a single `connectionString` — not both:

```yaml
databases:
  primary:
    type: postgres
    details:
      host: db.internal      # the DB address as seen from TableKit (or from the SSH host, if tunneling)
      port: 5432             # optional; per-engine default (pg 5432, mysql/mariadb 3306)
      database: app
      username: app_ro
      password: { from: env, env: PRIMARY_PW }
    tls:
      mode: verify-full      # disable (default) | allow | prefer | require | verify-ca | verify-full
      rootCertFilePath: /etc/ssl/db-ca.pem
  reporting:
    type: postgres
    connectionString: postgres://reader@warehouse.internal:5432/reporting
  legacy:
    type: mariadb
    details: { host: 10.0.0.5, username: reader }
    ssh:                     # optional: reach the DB through a bastion/jump host
      host: bastion.example.com
      username: deploy
      sshKeyFilePath: /keys/id_ed25519
```

Secrets (`password`, the SSH key `passphrase`) accept a bare string, or an
object: `{ from: env, env: VAR }`, `{ from: file, path: /run/secrets/x }`, or
`{ from: literal, value: ... }`. Every connection opens on its own, through its
own SSH tunnel and TLS settings when configured. The full field reference is the
JSON Schema at `core/engine/config/schemas/databases.schema.json`.

## How it works

TableKit listens on two ports. The **app** port (`8080`) serves the MCP endpoint
at `/mcp` and the OAuth endpoints next to it. The **control** port (`8081`) is
just `/health` and is meant for load balancers and uptime checks — keep it off
the public internet.

The MCP side speaks Streamable HTTP. Auth is plain OAuth 2.1: clients register
themselves dynamically and use PKCE, so there are no secrets to manage by hand.
Access is gated by pairing rather than a user database, which suits a server
that's yours alone. State — registered clients and pairing, refresh-token chains
and CLI bearer tokens, and stored queries — lives in a Postgres database of
TableKit's own, whose schema is brought up to date automatically on every start
(embedded goose migrations). The JWT signing key is not stored: it is supplied
via the required `SIGNING_KEY` env, so TableKit keeps nothing on local disk.

Each query runs on its own connection, reaching the database directly or through
a per-database SSH tunnel and/or TLS when configured. Read-only is enforced where
it counts: every query runs inside a read-only transaction, so TableKit won't run
writes or DDL on your behalf.

TableKit stores queries, not result rows. `run_query` runs read-only SQL, saves
the query (database, SQL, a description) to its own Postgres database, and returns
a `result_key`. Everything downstream takes that key and re-runs the stored SQL
against live data: `retrieve_results` pages through rows, `render_cartesian_series_chart`
and `render_proportional_chart` draw the in-chat MCP Apps charts (fed by an
app-only `fetch_chart_data`), and `get_export_url` mints a short-lived signed URL
the chart widget opens in your browser to download the full result as CSV or JSON.

## Pairing

```bash
tablekit pairing enable --once          # admit the next client, then lock again
tablekit pairing enable --indefinitely  # leave pairing open (e.g. while testing)
tablekit pairing disable                # turn new pairings off
```

Use `--once` for the normal case: open the door for one client, and it closes
behind them. `--indefinitely` is handy while you're wiring things up or want both
ChatGPT and Claude attached. Already-paired clients keep working regardless.

For a headless or scripted client that can't do the OAuth dance, mint a
long-lived bearer token instead:

```bash
tablekit pairing token:generate          # prints a tablekit_pat_… token and its id
tablekit pairing token:revoke <id OR token>
```

The client then presents it as `Authorization: Bearer <token>` to `/mcp`, no
pairing or OAuth flow needed. Revoke it any time by id or by the token itself.

## Development

TableKit is Go — [Gin](https://gin-gonic.com) for HTTP, [Cobra](https://cobra.dev)
for the CLI, and the official [MCP go-sdk](https://github.com/modelcontextprotocol/go-sdk).
`docker compose up` runs it with [Air](https://github.com/air-verse/air), so
editing any `.go` file under `core/` rebuilds and restarts the server in place.

```bash
go test ./...        # unit + e2e suite (DB container tests skip without Docker)
go test -race ./...  # the pairing path is concurrency-sensitive
```

The database e2e tests (`run_query` against real Postgres/MySQL, including over an
SSH tunnel and TLS) need Docker and the shared `tablekit` network, which exist
inside the `core` container — so run them through Compose:

```bash
docker compose up -d core
docker compose exec core go test ./e2e/...   # full DB + tunnel + TLS matrix
```

Outside that environment they `t.Skip` themselves, so a plain `go test ./...` on
your host stays green.

### Regenerating the state DB code

TableKit's own state is read and written through [go-jet](https://github.com/go-jet/jet),
whose typed table/model code (committed under `core/db/gen/`) is **generated from
the live schema** — go-jet cannot read migration SQL. After changing a migration:

```bash
docker compose up -d           # core applies the new migration to the dev postgres
docker compose restart core    # (or restart) so goose runs the migration
docker compose exec core sh -c \
  'JET_DSN=postgres://postgres:pw@postgres:5432/tablekit?sslmode=disable go run ./db/gen'
git add core/db/gen            # commit the regenerated code
```

`go run ./db/gen` defaults `JET_DSN` to the dev postgres on its host port
(`localhost:5433`); override it as above when running inside the `core` container.

### Running e2e tests (throw-away database containers)

The e2e suite spins up disposable Postgres/MySQL containers to test against real
engines, then tears them down. It does this by driving the **host's** Docker
daemon from inside the `core` container (Docker-outside-of-Docker): the `core`
image ships the `docker` CLI, and `docker-compose.yml` bind-mounts the host
socket (`/var/run/docker.sock`) into it. Containers it starts are therefore
plain siblings on your host daemon — you can see them with `docker ps` and, if a
test crashes mid-run, clean them up with `docker rm -f` on the host.

Database containers are attached to the stable `tablekit` network (exposed to
the suite as `TABLEKIT_E2E_DOCKER_NETWORK`), which `core` is also on, so the
test reaches a database by container name over that network — no port
publishing:

```bash
# what the suite does, roughly:
docker run -d --network tablekit --name testdb-<rand> -e POSTGRES_PASSWORD=pw postgres:17
# … core connects to  testdb-<rand>:5432  …
docker rm -f testdb-<rand>
```

Sanity-check the wiring once the stack is up:

```bash
docker compose exec core docker version       # reports a *Server* version → socket reaches the host daemon
docker compose exec core docker run --rm hello-world
```

#### Per-OS setup

- **Linux** — works out of the box: the socket is at `/var/run/docker.sock` and
  the container runs as root (so it can use the root-owned socket).
  `host.docker.internal` needs Docker Engine ≥ 20.10. Running **rootless
  Docker**? Your socket lives at `$XDG_RUNTIME_DIR/docker.sock` — point the bind
  mount's `source` there instead.
- **macOS (Docker Desktop)** — enable **Settings → Advanced → "Allow the default
  Docker socket to be used"** so `/var/run/docker.sock` exists on the host for
  the bind mount. `host.docker.internal` resolves natively.
- **Windows (Docker Desktop + WSL2)** — use the **WSL2 backend** and run
  `docker compose` **from inside a WSL2 distro** (not PowerShell/CMD) so the
  Linux socket path `/var/run/docker.sock` is valid. Enable the same "default
  Docker socket" setting. Native Windows containers / the `npipe` socket are not
  supported by this setup.

If a test can't use the shared network, the fallback is to publish the DB port
(`-p`) and connect via `host.docker.internal:<port>` (the `host-gateway`
mapping in `docker-compose.yml` makes that name resolve on Linux too).

```
core/
├── cli/                # tablekit CLI — serve, pairing (incl. bearer tokens)
├── db/                 # Postgres store (pgx stdlib) + embedded goose migrations
├── engine/             # read-only SQL engine — consumers only touch Service
│   ├── config/         # databases.yaml loading, validation, secret resolution
│   ├── driver/         # per-engine impls — postgres, mysql (also mariadb)
│   ├── transport/      # connection mechanics — sshtunnel, dbtls
│   └── encoding/       # row normalization + result shaping
├── mcp/                # the MCP server (wired to the services)
│   ├── handlers/       # the tools — one per file
│   └── ui/             # embedded MCP Apps widget builds (widgets/)
├── services/           # shared dependencies bundle
│   ├── config/         # environment config
│   ├── store/          # Postgres state (clients, pairing, tokens) + signing key file
│   ├── queries/        # stored-query repository (mcp_queries) over Postgres
│   ├── oauth/          # OAuth 2.1 issuer — JWT, PKCE, bearer + export tokens
│   └── services.go     # the Services bundle (config, store, engine, issuer, db, queries)
└── http/               # the two Gin listeners
    ├── app/            # public engine
    │   ├── oauth/      # OAuth 2.1 handlers — register, authorize, token, metadata
    │   │   └── templates/  # embedded HTML (already-paired page)
    │   ├── exports/    # signed CSV/JSON export endpoint (/exports/:format/:token)
    │   └── mcp.go      # mounts the MCP server on /mcp behind the bearer guard
    ├── commons/        # shared HTTP bits — the welcome page
    └── control/        # control engine — root, health
```

The in-chat charts live in `widgets/` — a separate React + Vite workspace
(shadcn/ui on Base UI + Tailwind, Recharts, the `@modelcontextprotocol/ext-apps`
guest SDK) built into single-file HTML and embedded into the binary. In dev the
`widgets` Compose service rebuilds them on change, which Air picks up to re-embed.

## Roadmap

Things we want next:

- More engines — SQLite and SQL Server alongside Postgres and MySQL.
- A few more chart types, and saved views you can re-open.
- An opt-in write mode, off by default and gated behind an explicit flag, for
  the cases where you really do want the assistant to make a change.
