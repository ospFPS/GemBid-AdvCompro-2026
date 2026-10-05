#!/usr/bin/env bash
# Deploy GemBid on the server: pull, rebuild, restart, health check, ensure ngrok.
#   ./deploy.sh                      deploy current branch
#   ./deploy.sh backend/db/migrations/019_x.sql   also apply that migration first
# Logs: docker compose logs -f backend
set -euo pipefail
cd "$(dirname "$0")"

git pull --ff-only

for migration in "$@"; do
    echo "Applying $migration"
    docker run --rm --env-file .env -v "$PWD/$(dirname "$migration")":/m:ro postgres:17-alpine \
        sh -c "psql \"\$DATABASE_URL\" -v ON_ERROR_STOP=1 -q -f /m/$(basename "$migration")"
done

docker compose up -d --build

timeout 30 sh -c 'until curl -fsS http://localhost:8080/health >/dev/null 2>&1; do sleep 1; done' \
    || { docker compose logs --tail 50 backend; exit 1; }
echo "Backend healthy on :8080 ($(git branch --show-current) @ $(git rev-parse --short HEAD))"

if pgrep -x ngrok >/dev/null; then
    echo "ngrok already running"
else
    nohup ngrok start --all --config="$HOME/.config/ngrok/ngrok.yml" >/tmp/ngrok.log 2>&1 &
    echo "ngrok started (log: /tmp/ngrok.log)"
fi
