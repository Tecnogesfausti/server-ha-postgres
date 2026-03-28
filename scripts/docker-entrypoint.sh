#!/bin/sh
# docker-entrypoint.sh

# Immediately exit if any command has a non-zero exit status.
set -e

# Execute DB migrations only when the main application is about to run
if [ $# -eq 1 ] && [ "${1:-}" = "/app/app" ]; then
  if [ "${DATABASE__DIALECT:-}" = "postgres" ] || grep -q 'dialect:[[:space:]]*postgres' /app/config.yml 2>/dev/null; then
    /app/app db:auto-migrate
  else
    /app/app db:migrate up
  fi
fi

# Execute the main application
exec "$@"
