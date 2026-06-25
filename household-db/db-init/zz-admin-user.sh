#!/bin/bash
set -euo pipefail

# Inserts an application user (into the `users` table) using credentials from
# the environment. The password is bcrypt-hashed via pgcrypto so it is
# compatible with Go's bcrypt verification in internal/login.
#
# Runs after init.sql (alphabetical order) so the `users` table already exists.
# Like all init scripts, it only runs on first container start (empty volume).

: "${ADMIN_USERNAME:?ADMIN_USERNAME is not set}"
: "${ADMIN_PASSWORD:?ADMIN_PASSWORD is not set}"

psql -v ON_ERROR_STOP=1 \
     --username "$POSTGRES_USER" \
     --dbname "$POSTGRES_DB" \
     --set=admin_user="$ADMIN_USERNAME" \
     --set=admin_pwd="$ADMIN_PASSWORD" <<'EOSQL'
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (username, pwd)
VALUES (:'admin_user', crypt(:'admin_pwd', gen_salt('bf', 12)))
ON CONFLICT (username) DO NOTHING;
EOSQL
