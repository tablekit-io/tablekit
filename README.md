# tablekit

Minimal Go service built with [Fiber](https://gofiber.io) and a [Cobra](https://cobra.dev) CLI, with live-reload via [Air](https://github.com/air-verse/air), all orchestrated through Docker Compose.

## Layout

```
.
├── docker-compose.yml      # runs the core service
└── core/
    ├── main.go             # Cobra CLI + Fiber server
    ├── go.mod / go.sum
    ├── .air.toml           # live-reload config (builds /build/tablekit, runs `serve`)
    ├── Dockerfile          # golang base + Air
    ├── .dockerignore
    └── .gitignore
```

## Run

```bash
docker compose up --build
```

Then:

```bash
curl localhost:8080
# hello world
```

Edit any `.go` file in `core/` — Air rebuilds and restarts the server inside the container automatically.

Stop:

```bash
docker compose down
```

## CLI

The binary is `tablekit`. The HTTP server only starts under the `serve` subcommand:

```bash
tablekit serve     # start the Fiber server on :8080
tablekit           # print usage (no server)
```

Inside the running container:

```bash
docker compose exec core /build/tablekit serve
```

## Notes

- Air builds to `/build` (container-only), so no build artifacts land in the host `core/` directory.
- Dependencies resolve at container start via `go mod tidy` (see `core/Dockerfile`).
