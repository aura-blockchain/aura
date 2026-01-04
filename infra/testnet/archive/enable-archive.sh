#!/usr/bin/env bash
set -euo pipefail

AURA_HOME="${AURA_HOME:-$HOME/.aura-archive}"
APP_CONFIG="${AURA_HOME}/config/app.toml"

if [ ! -f "${APP_CONFIG}" ]; then
  echo "Missing app.toml at ${APP_CONFIG}"
  exit 1
fi

sed -i.bak -E 's|^pruning *=.*|pruning = "nothing"|' "${APP_CONFIG}"
sed -i.bak -E 's|^pruning-keep-recent *=.*|pruning-keep-recent = "0"|' "${APP_CONFIG}"
sed -i.bak -E 's|^pruning-keep-every *=.*|pruning-keep-every = "0"|' "${APP_CONFIG}"
sed -i.bak -E 's|^pruning-interval *=.*|pruning-interval = "0"|' "${APP_CONFIG}"

echo "Archive settings applied at ${APP_CONFIG}"
