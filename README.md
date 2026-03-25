# gitea-hooks

Automated PR code review using Claude Code. Listens for Gitea webhooks and posts AI-powered review comments.

## Features

- Receives Gitea PR webhooks
- Validates webhook signatures (HMAC-SHA256)
- Clones PR branches via SSH
- Runs Claude Code CLI to review code
- Posts review as PR comment via Gitea API
- Sequential job processing with configurable queue
- Automatic cleanup of cloned repos
- Runs in Docker

## Quick Start

1. **Copy environment config**
   ```bash
   cp .env.example .env
   ```

2. **Configure `.env`**
   ```bash
   GITEA_URL=https://your-gitea.com
   GITEA_TOKEN=your_api_token
   WEBHOOK_SECRET=your_webhook_secret
   SSH_KEY_HOST_PATH=~/.ssh/id_ed25519
   CLAUDE_SKILL=your-pr-review-skill
   ```

3. **Authenticate Claude CLI** (choose one)

   Option A - Pro subscription:
   ```bash
   claude login
   ```

   Option B - API key:
   ```bash
   # Add to .env
   ANTHROPIC_API_KEY=sk-ant-...
   ```

4. **Start the server**
   ```bash
   docker compose up -d
   ```

5. **Configure Gitea webhook**
   - Go to your repo Settings > Webhooks > Add Webhook
   - URL: `http://your-server:8080/webhook`
   - Secret: same as `WEBHOOK_SECRET`
   - Events: Pull Request

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITEA_URL` | Yes | - | Gitea instance URL |
| `GITEA_TOKEN` | Yes | - | API token for posting comments |
| `WEBHOOK_SECRET` | Yes | - | Shared secret for signature validation |
| `SSH_KEY_HOST_PATH` | Yes | - | Host path to SSH key (for Docker mount) |
| `SSH_KEY_PATH` | No | `/secrets/id_ed25519` | Container path to SSH key |
| `CLAUDE_SKILL` | Yes | - | Claude Code skill to run |
| `ANTHROPIC_API_KEY` | No* | - | API key for Claude (pay-per-use) |
| `CLAUDE_CONFIG_PATH` | No* | `~/.claude` | Path to Claude config (Pro subscription) |
| `PORT` | No | `8080` | Server port |
| `HOST_PORT` | No | `8080` | Docker host port |
| `CLONE_BASE_PATH` | No | `/data/reviews` | Where to clone repos |
| `RETENTION_HOURS` | No | `24` | Hours to keep cloned repos |
| `QUEUE_SIZE` | No | `100` | Max queued jobs |

*One of `ANTHROPIC_API_KEY` or `CLAUDE_CONFIG_PATH` is required.

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/webhook` | POST | Receives Gitea webhooks |
| `/health` | GET | Health check (returns "ok") |

## Development

```bash
# Run tests
go test ./... -v

# Run with race detector
go test ./... -race

# Build locally
go build -o gitea-hooks .

# Build Docker image
docker compose build

```

## Architecture

```
                    ┌─────────────┐
   Gitea Webhook───▶│   Server    │
                    │  /webhook   │
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │    Queue    │
                    │  (in-mem)   │
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   Worker    │
                    │             │
                    │ 1. Clone    │
                    │ 2. Review   │
                    │ 3. Comment  │
                    └─────────────┘
```

## License

MIT
