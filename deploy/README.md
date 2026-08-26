# 旅迹 · 部署说明

把整个 `deploy` 文件夹拷到 Ubuntu 24 服务器的 `/data/soft/travel/deploy/`。Go 进程只监听本机 `127.0.0.1:5000`，对外由 Caddy 提供 `https://travel.regen.ltd`。

## Ubuntu 24 一键部署

DNS：把 `travel.regen.ltd` 的 A 记录指到服务器公网 IP。不要开 CDN 小云朵（或 SSL 设为 Full），否则证书会失败。防火墙放行 **22 / 80 / 443**。

```
sudo mkdir -p /data/soft/travel
# 将 deploy 目录上传到 /data/soft/travel/deploy
cd /data/soft/travel/deploy
sudo bash install.sh
```

脚本会：安装 Caddy、创建 `travel` 用户、注册 systemd 服务、写入反代配置并启动。之后开机自启。

浏览器打开：https://travel.regen.ltd/

```
systemctl status travel
systemctl status caddy
journalctl -u travel -f
journalctl -u caddy -f
```

## 目录

```
deploy/
├── install.sh                 # Ubuntu 24 一键部署
├── travel-server.exe          # Windows 64 位
├── travel-server              # Linux amd64
├── dist/                      # 前端静态文件
├── data/                      # SQLite（首次启动自动创建）
├── uploads/                   # 用户上传的图片
├── .env                       # 运行配置（含密钥，勿公开）
├── Caddyfile                  # travel.regen.ltd 反代
├── CREDENTIALS.txt            # 初始账号
├── start.bat                  # Windows 手动启动
├── start.sh                   # Linux 前台启动（一般不用）
└── start-caddy.sh             # Linux 前台启动 Caddy（一般不用）
```

## 账号

见同目录 `CREDENTIALS.txt`。

- 不能自行注册
- 只有管理员可以添加用户（登录后右上角 → 用户管理）
- 测试账号 `demo` 已删除

## 上线注意

1. `.env` 中 `PUBLIC_URL` 已设为 `https://travel.regen.ltd`
2. `.env` 中 `LISTEN=127.0.0.1`，仅允许 Caddy 访问后端
3. 妥善保存 `CREDENTIALS.txt`，不要发到公开仓库

## 重新打包

在项目根目录执行：

```
powershell -File scripts/pack-deploy.ps1
```
