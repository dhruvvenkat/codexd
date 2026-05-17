# codexd

`codexd` is a small local web GUI for controlling Codex CLI sessions running inside `tmux`.

It is intentionally local-first: the daemon binds to `127.0.0.1:7777` by default, uses a static bearer token, and does not expose arbitrary shell execution.

## Requirements

- Linux
- Go 1.22+
- Node.js and npm for building the web UI
- `tmux`
- `git`
- `codex`

## Run

Build the frontend first:

```sh
cd web
npm install
npm run build
cd ..
```

Start the daemon:

```sh
go run ./cmd/codexd
```

On first run, `codexd` creates:

- Config: `~/.config/codexd/config.json`
- Token: `~/.config/codexd/token`
- State: `~/.local/share/codexd/state.json`

Open `http://127.0.0.1:7777` and paste the bearer token from the token file.

## Development

Run the backend:

```sh
go run ./cmd/codexd
```

Run the Vite dev server:

```sh
cd web
npm run dev
```

Vite proxies `/api` to `http://127.0.0.1:7777`.

## API

All API routes require:

```text
Authorization: Bearer <token>
```

except:

```text
GET /api/health
```

Routes:

```text
GET    /api/health
GET    /api/sessions
POST   /api/sessions
GET    /api/sessions/:id
GET    /api/sessions/:id/output
POST   /api/sessions/:id/input
GET    /api/sessions/:id/git/status
GET    /api/sessions/:id/git/diff
DELETE /api/sessions/:id
```

Create a session:

```sh
TOKEN="$(cat ~/.config/codexd/token)"
curl -sS \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"codexd","repoPath":"/home/dhruv/Documents/Programming/codexd"}' \
  http://127.0.0.1:7777/api/sessions
```

Send input:

```sh
curl -sS \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"fix the failing tests and explain what changed"}' \
  http://127.0.0.1:7777/api/sessions/codexd/input
```

## Security Notes

`codexd` can send input to a live terminal session. Keep it private.

- It binds to localhost by default.
- It uses bearer token auth for API routes.
- It does not expose `/exec`, `/run`, `/shell`, upload, file edit, or environment routes.
- It does not log bearer tokens or request bodies.
- Remote access should use Tailscale, an SSH tunnel, or another trusted private network layer.
- Do not expose `codexd` directly to the public internet.

## Test

```sh
go test ./...
cd web && npm run build
```
