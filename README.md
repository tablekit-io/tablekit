# TableKit

Talk to your database from ChatGPT and Claude. TableKit is a small,
self-hostable MCP server you point at your Postgres and MySQL databases — then
ask questions in plain English and get back tables and interactive charts. Every
query it runs is read-only, so the assistant can look but never touch.

You run it on your own infrastructure. Your connection strings and data stay
with you; the only thing the assistant ever sees is the rows your read-only
queries return.

## What it does

- Reads your schema so the assistant knows your tables, columns and relationships.
- Turns questions into read-only SQL — `SELECT`s only, no writes, no DDL.
- Renders results as interactive charts right inside the chat, using MCP Apps
  (you can sort, hover, and switch chart types without leaving the conversation).
- Holds several named database connections at once, so you can keep `analytics`
  and `billing` side by side and tell the assistant which one to use.
- Works with ChatGPT on the web and Claude on web and desktop.

<!-- placeholder: replace with a real screenshot once captured -->
![A chart rendered from a query inside the chat](docs/chart.png)

## Quickstart

You'll need Docker. Point TableKit at one or more databases, give it the URL it
will be reachable on, and bring it up:

```bash
# .env
PUBLIC_BASE_URL=https://tablekit.your-host.com

# one entry per database — the part after TABLEKIT_DB_ is its name
TABLEKIT_DB_ANALYTICS=postgres://readonly:pw@db-host:5432/analytics
TABLEKIT_DB_BILLING=mysql://readonly:pw@db-host:3306/billing
```

```bash
docker compose up --build
```

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

| Variable            | Default                  | What it's for                                   |
|---------------------|--------------------------|-------------------------------------------------|
| `TABLEKIT_DB_<NAME>`| —                        | A database connection. Repeat it, one per DB.   |
| `PUBLIC_BASE_URL`   | `http://localhost:8080`  | The URL clients reach TableKit on.              |
| `APP_PORT`          | `8080`                   | MCP + OAuth listener.                           |
| `CONTROL_PORT`      | `8081`                   | Health and ops listener.                        |
| `DATA_DIR`          | `./data`                 | Where pairing + token state is kept.            |
| `SIGNING_KEY`       | generated                | Base64 HS256 key. Set it to share one key across instances; otherwise one is generated under `DATA_DIR`. Short keys are zero-padded to 32 bytes. |
| `ACCESS_TTL`        | `15m`                    | Access token lifetime.                          |
| `REFRESH_TTL`       | `168h`                   | Refresh token lifetime.                         |

## How it works

TableKit listens on two ports. The **app** port (`8080`) serves the MCP endpoint
at `/mcp` and the OAuth endpoints next to it. The **control** port (`8081`) is
just `/health` and is meant for load balancers and uptime checks — keep it off
the public internet.

The MCP side speaks Streamable HTTP. Auth is plain OAuth 2.1: clients register
themselves dynamically and use PKCE, so there are no secrets to manage by hand.
Access is gated by pairing rather than a user database, which suits a server
that's yours alone. State — the paired client, token chains, and the signing key
— lives as JSON files under `DATA_DIR`, generated on first boot and gitignored.

Read-only is enforced where it counts: TableKit will not emit anything but
`SELECT`, and it won't run DDL or DML on your behalf.

## Pairing

```bash
tablekit pairing enable --once          # admit the next client, then lock again
tablekit pairing enable --indefinitely  # leave pairing open (e.g. while testing)
tablekit pairing disable                # turn new pairings off
```

Use `--once` for the normal case: open the door for one client, and it closes
behind them. `--indefinitely` is handy while you're wiring things up or want both
ChatGPT and Claude attached. Already-paired clients keep working regardless.

## Development

TableKit is Go — [Gin](https://gin-gonic.com) for HTTP, [Cobra](https://cobra.dev)
for the CLI, and the official [MCP go-sdk](https://github.com/modelcontextprotocol/go-sdk).
`docker compose up` runs it with [Air](https://github.com/air-verse/air), so
editing any `.go` file under `core/` rebuilds and restarts the server in place.

```bash
go test ./...        # unit + e2e suite
go test -race ./...  # the pairing path is concurrency-sensitive
```

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
docker run -d --network tablekit --name testdb-<rand> -e POSTGRES_PASSWORD=pw postgres:16
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
├── cli/                # tablekit CLI — serve, pairing
├── mcp/                # the MCP server
│   ├── handlers/       # the tools — one per file
│   └── ui/             # embedded MCP Apps widget builds (dist)
├── services/           # shared dependencies
│   ├── config/         # environment config
│   ├── store/          # JSON state (connections, pairing, tokens, signing key)
│   └── services.go     # the Services bundle (config + store)
└── http/               # the two Gin listeners
    ├── app/            # public engine
    │   ├── oauth/      # OAuth 2.1 — register, authorize, token, metadata
    │   └── mcp.go      # mounts the MCP server on /mcp behind the bearer guard
    └── control/        # control engine — root, health
```

## Roadmap

Things we want next:

- More engines — SQLite and SQL Server alongside Postgres and MySQL.
- A few more chart types, and saved views you can re-open.
- An opt-in write mode, off by default and gated behind an explicit flag, for
  the cases where you really do want the assistant to make a change.
