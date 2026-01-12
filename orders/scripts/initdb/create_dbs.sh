#!/usr/bin/env sh
set -e

# Create extra databases on first initialization.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'SQL'
CREATE DATABASE orders_db;
CREATE DATABASE payment;
SQL
