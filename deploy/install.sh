#!/bin/bash
# 旅迹 · Ubuntu 24 一键部署
# 用法（在本目录）：
#   sudo bash install.sh
#
# 预期路径：/data/soft/travel/deploy

if [ -z "${TRAVEL_INSTALL_CRLF_FIXED:-}" ]; then
  APP_DIR="$(cd "$(dirname "$0")" && pwd)"
  export TRAVEL_APP_DIR="$APP_DIR"
  export TRAVEL_INSTALL_CRLF_FIXED=1
  tmp="$(mktemp)"
  tr -d '\r' < "$0" > "$tmp"
  exec bash "$tmp" "$@"
fi

APP_DIR="${TRAVEL_APP_DIR:-$(cd "$(dirname "$0")" && pwd)}"
set -euo pipefail

DOMAIN="${TRAVEL_DOMAIN:-travel.regen.ltd}"
SERVICE_NAME="travel"
APP_USER="travel"
LISTEN_PORT="5000"

log() { echo "[travel] $*"; }
die() { echo "[travel] 错误: $*" >&2; exit 1; }

if [ "$(id -u)" -ne 0 ]; then
  die "请用 root 运行：sudo bash install.sh"
fi

export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=l

for f in install.sh start.sh start-caddy.sh Caddyfile; do
  if [ -f "$APP_DIR/$f" ]; then
    sed -i 's/\r$//' "$APP_DIR/$f" || true
  fi
done

[ -x "$APP_DIR/travel-server" ] || [ -f "$APP_DIR/travel-server" ] || die "找不到 $APP_DIR/travel-server（Linux 可执行文件）"
[ -f "$APP_DIR/.env" ] || die "找不到 $APP_DIR/.env"
[ -f "$APP_DIR/dist/index.html" ] || die "找不到前端 dist/index.html"
[ -f "$APP_DIR/Caddyfile" ] || die "找不到 Caddyfile"

if [ -f /etc/os-release ]; then
  # shellcheck disable=SC1091
  . /etc/os-release
  log "系统: ${PRETTY_NAME:-unknown}"
fi

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) ;;
  *) die "当前架构 $arch 不受支持，需要 x86_64 的 travel-server" ;;
esac

mkdir -p "$APP_DIR/data" "$APP_DIR/uploads"
chmod +x "$APP_DIR/travel-server"

if ! id -u "$APP_USER" >/dev/null 2>&1; then
  log "创建系统用户 $APP_USER"
  useradd --system --home "$(dirname "$APP_DIR")" --shell /usr/sbin/nologin "$APP_USER"
fi

chown -R "$APP_USER:$APP_USER" "$APP_DIR"
chmod 750 "$APP_DIR"
chmod 640 "$APP_DIR/.env"
chmod 755 "$APP_DIR/travel-server"

log "安装依赖"
apt-get update -qq
apt-get install -y -qq ca-certificates curl gnupg debian-keyring debian-archive-keyring apt-transport-https

install_caddy_apt() {
  mkdir -p /usr/share/keyrings
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
    | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
    | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
  apt-get update -qq
  apt-get install -y -qq caddy
}

install_caddy_static() {
  log "改从 GitHub 安装 Caddy 二进制"
  if ! id -u caddy >/dev/null 2>&1; then
    useradd --system --home /var/lib/caddy --create-home --shell /usr/sbin/nologin caddy
  fi
  mkdir -p /usr/bin /etc/caddy /var/lib/caddy
  chown -R caddy:caddy /var/lib/caddy
  tmpd="$(mktemp -d)"
  trap 'rm -rf "$tmpd"' RETURN
  ver="$(curl -fsSL https://api.github.com/repos/caddyserver/caddy/releases/latest | grep -o '"tag_name": *"[^"]*"' | head -n1 | cut -d'"' -f4 || true)"
  ver="${ver:-v2.10.0}"
  url="https://github.com/caddyserver/caddy/releases/download/${ver}/caddy_${ver#v}_linux_amd64.tar.gz"
  log "下载 $url"
  curl -fL "$url" -o "$tmpd/caddy.tgz"
  tar -xzf "$tmpd/caddy.tgz" -C "$tmpd" caddy
  install -m 755 "$tmpd/caddy" /usr/bin/caddy
  cat > /etc/systemd/system/caddy.service <<'UNIT'
[Unit]
Description=Caddy
Documentation=https://caddyserver.com/docs/
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
ExecStart=/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile
ExecReload=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force
TimeoutStopSec=5s
LimitNOFILE=1048576
PrivateTmp=true
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
UNIT
}

if command -v caddy >/dev/null 2>&1; then
  log "已安装 Caddy $(caddy version 2>/dev/null | head -n1)"
elif install_caddy_apt; then
  log "已用 apt 安装 Caddy $(caddy version 2>/dev/null | head -n1)"
elif install_caddy_static; then
  log "已用静态包安装 Caddy $(caddy version 2>/dev/null | head -n1)"
else
  die "Caddy 安装失败，请检查网络后重试"
fi

command -v caddy >/dev/null 2>&1 || die "Caddy 未出现在 PATH 中"

log "写入 systemd: ${SERVICE_NAME}.service"
cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=Travel Footprints (旅迹)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_USER}
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/travel-server
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

log "写入 /etc/caddy/Caddyfile"
cp "$APP_DIR/Caddyfile" /etc/caddy/Caddyfile
if getent group caddy >/dev/null 2>&1; then
  chown root:caddy /etc/caddy/Caddyfile
else
  chown root:root /etc/caddy/Caddyfile
fi
chmod 644 /etc/caddy/Caddyfile

caddy validate --config /etc/caddy/Caddyfile >/dev/null

if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -qi 'Status: active'; then
  log "UFW 已启用，放行 80/443（保留 OpenSSH）"
  ufw allow OpenSSH >/dev/null
  ufw allow 80/tcp >/dev/null
  ufw allow 443/tcp >/dev/null
fi

systemctl daemon-reload
systemctl enable "${SERVICE_NAME}.service" >/dev/null
systemctl restart "${SERVICE_NAME}.service"
systemctl enable caddy >/dev/null
systemctl restart caddy

log "等待应用就绪"
ok=0
for _ in $(seq 1 30); do
  if curl -sf "http://127.0.0.1:${LISTEN_PORT}/api/health" >/dev/null; then
    ok=1
    break
  fi
  sleep 0.4
done
[ "$ok" = 1 ] || die "应用未在 127.0.0.1:${LISTEN_PORT} 起来，请查看: journalctl -u ${SERVICE_NAME} -e"

systemctl is-active --quiet caddy || die "Caddy 未运行，请查看: journalctl -u caddy -e"

echo
log "部署完成"
echo "  应用目录  $APP_DIR"
echo "  本机接口  http://127.0.0.1:${LISTEN_PORT}/api/health"
echo "  公网站点  https://${DOMAIN}/"
echo
echo "  应用状态  systemctl status ${SERVICE_NAME}"
echo "  Caddy状态 systemctl status caddy"
echo "  应用日志  journalctl -u ${SERVICE_NAME} -f"
echo "  Caddy日志 journalctl -u caddy -f"
echo
echo "请确认 ${DOMAIN} 的 A 记录已指向本机公网 IP，否则证书申请会失败。"
echo "账号见 $APP_DIR/CREDENTIALS.txt"
