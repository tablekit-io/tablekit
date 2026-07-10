# TableKit

[![GitHub License](https://img.shields.io/github/license/tablekit-io/tablekit)
](https://github.com/tablekit-io/tablekit/blob/main/LICENSE) [![GitHub last commit](https://img.shields.io/github/last-commit/tablekit-io/tablekit)
](https://github.com/tablekit-io/tablekit) [![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/tablekit-io/tablekit)
](https://github.com/tablekit-io/tablekit/issues) 

MCP server that lets you ask ChatGPT or Claude real data questions to get analytics, charts and visualizations straight in your chat. 

**Database Support**: `Postgres`, `MySQL`, `BigQuery` (coming soon)



https://github.com/user-attachments/assets/424b887a-3f22-47ff-9df6-2a3b2837e86a





## Features

1. 📡 **Remote MCP** — HTTP with OAuth authentication. Basically a fancy way to say it works with Claude web or ChatGPT web.
2. 📈 **Interactive Charts Within Chat** — Based on the new [MCP Apps](https://modelcontextprotocol.io/extensions/apps/overview) spec. View data or review SQL.
3. 💽 **JSON / CSV Export** — from withint he interactive chart widget.
4. 💯 **Multiple Databases Support** — Connect more than one database (Postgres / MySQL).
5. 🕳️ **SSH Tunnel Support** — Reach databases behind private networks.
6. 🤖 **Multiple Client Support** — Connect more than one ChatGPT / Claude account to the same TableKit server. 
7. 🔐 **Read-only Queries** — ensured via read-only transactions.

### Supported Chart Types

1. Bar Chart
2. Line Chart / Area Chart
3. Pie Chart / Donut Chart
4. Sunburst Chart
5. `COMING SOON` Cohort Chart / Table Chart
6. `COMING SOON` Gauge / Progress
7. `COMING SOON` Number / Stat

## Quickstart

### Clone + Copy Config Files

There isn't an installation step per se, you just clone and run it via Docker Compose.

```bash
# Clone TableKit
git clone https://github.com/tablekit-io/tablekit.git

cd tablekit

# Copy config file
cp docker-compose.example.yml docker-compose.yml
cp databases.example.yml databases.yml
```

### Create Public URL

TableKit server by default starts on port `8080`.

If you have your own domain, use a reverse proxy, something like [nginx](https://nginx.org/) or [traefik](https://traefik.io/traefik), coupled with a [certbot](https://certbot.eff.org/) plugin to generate a public TLS terminated base URL for TableKit.

Alternatively you could use something like [ngrok](https://ngrok.com/), or [Cloudflare Tunnels](https://developers.cloudflare.com/tunnel/setup/#quick-tunnels-development) to tunnel your local infra on to the public internet. (this will allow ngrok or cloudflare to intercept your data, use with caution).

We will need this URL in the next step to allow clients (like ChatGPT and Claude) to complete the OAuth flow with TableKit.

### Create `.env`

```bash
# generate signing key
openssl rand -base64 32

# generate a secure password for db
openssl rand -base64 64 | tr -dc 'a-zA-Z0-9' | head -c 32; echo
```

Create a `.env` file in your tablekit directory, with the following contents. Replace the values from previous steps.

```bash
PUBLIC_BASE_URL= # insert public base url
DB_PASSWORD= # insert generated db password
SIGNING_KEY= # insert generated signing key

# this helps us immensely
ENABLE_ANONYMOUS_TELEMETRY=1
```

### Edit `databases.yml`

With your database's details of course.

```yaml
databases:
  name-your-db:
    type: postgres
    details:
      host: postgres.server-address.com
      port: 5432
      database: your-app
      username: postgres
      password: postgres
```

### Start TableKit Server

```bash
docker compose up --build -d
```

### Add To ChatGPT / Claude

#### ChatGPT

1. Go to [Security and login](https://chatgpt.com/plugins#settings/Security:~:text=Developer%20mode-,Developer%20mode,-ELEVATED%20RISK) settings.
2. Enable `Developer mode`
3. Go to [New App](https://chatgpt.com/plugins#settings/Connectors?create-connector=true&redirectAfter=%2Fplugins)
4. Enter `TableKit` in name
5. Set the _Server URL_ to `https://[PUBLIC BASE URL]/mcp`
6. Check `I understand and want to continue` at the end
7. Click **Create**

#### Claude

1. Go to [Add Custom Connector](https://claude.ai/chat/022f63df-c2cc-405b-842c-38e7bc0ecf3e?modal=add-custom-connector#settings/customize-connectors)
2. Enter `TableKit` in the name
3. Enter `https://[PUBLIC BASE URL]/mcp` in the "Remote MCP server URL"
4. Click **Add**

### Done 🎉

Try TableKit with prompts like

```
Can you see connected databases in TableKit?
```

```
How many users do we have in our database?
```

```
Show me a week over week histogram of user growth based on our db
```


# License

MIT
