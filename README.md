# tablekit

Single-client **OAuth 2.1 HTTP MCP server** in Go, built with [Gin](https://gin-gonic.com)
and a [Cobra](https://cobra.dev) CLI, using the official
[MCP go-sdk](https://github.com/modelcontextprotocol/go-sdk). Live-reload via
[Air](https://github.com/air-verse/air), orchestrated through Docker Compose.

It exposes one `hello_world` tool over Streamable HTTP, protected by an OAuth
flow (dynamic client registration, PKCE, JWT access/refresh tokens). All state
lives in gitignored JSON files — no database.

## Layout

```
.
├── docker-compose.yml      # runs the core service (two published ports)
└── core/
    ├── main.go             # Cobra CLI entrypoint
    ├── cmd/                # serve command (starts both listeners)
    ├── config/             # env-driven config (ports, base URL, TTLs)
    ├── store/              # JSON persistence (clients, tokens, signing key)
    ├── oauth/              # OAuth 2.1: register, authorize, token, metadata, JWT
    ├── mcpserver/          # go-sdk MCP server + hello_world + bearer guard
    ├── server/            # Gin engines (app + control) and handlers
    ├── go.mod / go.sum
    ├── .air.toml           # live-reload config (builds /build/tablekit, runs `serve`)
    ├── Dockerfile          # golang base + Air
    └── .gitignore          # ignores data/ (OAuth state)
```

## Ports

| Port  | Purpose                                                                 |
|-------|-------------------------------------------------------------------------|
| 8080  | **app** — MCP (`/mcp`) + OAuth (`/oauth/*`, `/register`, `/.well-known/*`) |
| 8081  | **control** — `/`, `/health`, reserved for ops                          |

All ports and lifetimes are env-overridable: `APP_PORT`, `CONTROL_PORT`,
`PUBLIC_BASE_URL`, `DATA_DIR`, `ACCESS_TTL` (default `15m`), `REFRESH_TTL`
(default `168h`).

## Run

```bash
docker compose up --build
```

Then:

```bash
curl localhost:8081/health
# {"status":"OK", ...}

curl localhost:8080/.well-known/oauth-authorization-server
# OAuth discovery document
```

Edit any `.go` file in `core/` — Air rebuilds and restarts the server inside the container automatically.

Stop:

```bash
docker compose down
```

## Connecting an MCP client

Point an OAuth-capable MCP client at `http://localhost:8080/mcp`. It will
discover the auth server, register dynamically, and run the PKCE flow. The
**first** client to complete authorization becomes the paired client; any other
client is told `already paired`. Refresh tokens rotate on use — replaying a
superseded refresh token revokes the whole chain.

## CLI

The binary is `tablekit`. The HTTP server only starts under the `serve` subcommand:

```bash
tablekit serve     # start the app (:8080) and control (:8081) listeners
tablekit           # print usage (no server)
```

Inside the running container:

```bash
docker compose exec core /build/tablekit serve
```

## Notes

- Air builds to `/build` (container-only), so no build artifacts land in the host `core/` directory.
- Dependencies resolve at container start via `go mod tidy` (see `core/Dockerfile`).
- OAuth state (`core/data/`: `clients.json`, `tokens.json`, `signing.key`) is
  gitignored. The signing key is generated on first boot.
