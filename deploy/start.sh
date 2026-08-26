#!/bin/bash
cd "$(dirname "$0")"
mkdir -p data uploads
chmod +x ./travel-server 2>/dev/null || true
echo "旅迹服务启动中（仅本机 127.0.0.1:5000，对外请走 Caddy）..."
exec ./travel-server
