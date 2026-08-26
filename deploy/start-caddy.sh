#!/bin/bash
cd "$(dirname "$0")"
if ! command -v caddy >/dev/null 2>&1; then
  echo "未找到 caddy，请先安装：https://caddyserver.com/docs/install"
  exit 1
fi
echo "Caddy 反代 https://travel.regen.ltd -> 127.0.0.1:5000"
exec caddy run --config ./Caddyfile --adapter caddyfile
