# 旅迹 · 旅行足迹地图

Vue 3 前端 + Go 后端的多用户旅行足迹服务。在地图上记录去过的地方，用照片轮播重温旅途，并通过链接与好友共享同一张地图。

---

## 功能

- 登录后使用；未开放自行注册，账号只能由管理员创建
- 个人头像与昵称，地图和共享页都会显示记录者
- 搜索地点即可添加足迹，也可以点击地图选点
- 分类、日期、旅行笔记、多图上传
- 详情页图片轮播，点击可全屏预览
- 时间线侧栏、搜索与分类筛选
- 创建共享地图：全部或挑选部分足迹，公开链接或仅邀请
- 按用户名邀请好友查看，也可一起在同一张地图上记录
- 公开分享页无需登录即可浏览

生产环境账号与启动方式见 `deploy/README.md` 和 `deploy/CREDENTIALS.txt`。

---

## 项目结构

```
vueTravelFootprints/
├── src/                         # Vue 3 前端
│   ├── api/                     # 接口封装（auth / markers / shares）
│   ├── assets/styles/           # 设计系统与全局样式
│   ├── components/
│   │   ├── common/              # 图片轮播等通用组件
│   │   ├── layout/              # 顶栏、侧栏
│   │   ├── map/                 # 地图与标记
│   │   ├── marker/              # 足迹详情、编辑表单
│   │   └── share/               # 分享对话框
│   ├── stores/                  # Pinia（auth / markers / ui）
│   ├── utils/                   # 分类、图片、日期工具
│   ├── views/                   # 登录页、地图页（含分享页）
│   ├── router/
│   └── main.js
├── backend-go/                  # Go 后端
│   ├── main.go
│   ├── config/
│   ├── models/                  # 用户、足迹、分享
│   ├── database/                # SQLite 与迁移
│   ├── handlers/                # HTTP 接口
│   ├── middleware/              # CORS、JWT
│   ├── data/travel.db
│   └── uploads/
├── assets/readmeImages/         # README 配图
├── index.html
└── package.json
```

---

生产部署请使用 `deploy/` 目录，说明见 [deploy/README.md](./deploy/README.md)。

## 快速开始

### 1. 环境变量

复制 `.env.eg` 为 `.env`：

```
VITE_AMAP_KEY=your_amap_key_here
VITE_AMAP_SECURITY_CODE=your_amap_security_js_code
VITE_API_BASE_URL=http://localhost:5000
AMAP_KEY=your_amap_web_key_or_js_key
ADMIN_USERNAME=admin
ADMIN_PASSWORD=
LIME_PASSWORD=
EIINXYZ_PASSWORD=
JWT_SECRET=change-me-in-production
```

`VITE_AMAP_SECURITY_CODE` 是高德控制台里的「安全密钥」，地点搜索需要它。后端会读取项目根目录或 `backend-go/` 下的 `.env`。

后端可选环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `5000` | 服务端口 |
| `LISTEN` | 空（所有网卡） | 监听地址；Caddy 反代时设为 `127.0.0.1` |
| `DB_PATH` | `./data/travel.db` | SQLite 路径 |
| `UPLOAD_DIR` | `./uploads` | 图片上传目录 |
| `JWT_SECRET` | 开发用默认值 | 生产环境务必修改 |
| `PUBLIC_URL` | 空 | 上传文件公网前缀；生产环境为 `https://travel.regen.ltd` |
| `AMAP_KEY` | 同 `VITE_AMAP_KEY` | 地点搜索 Web 服务 Key |
| `STATIC_DIR` | `./dist` | 前端静态目录，存在时由后端一并托管 |
| `ADMIN_USERNAME` | `admin` | 初始管理员用户名 |
| `ADMIN_PASSWORD` | 空 | 首次创建管理员时必填，至少 10 位 |
| `LIME_PASSWORD` | 空 | 若填写则创建普通账号 `lime` |
| `EIINXYZ_PASSWORD` | 空 | 若填写则创建普通账号 `eiinxyz` |

### 2. 启动后端

```
cd backend-go
go run .
```

服务默认运行在 `http://localhost:5000`。

### 3. 启动前端

```
npm install
npm run dev
```

生产构建：

```
npm run build
```

---

## 主要接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/auth/login` | 登录（用户名或邮箱） |
| `GET` | `/api/auth/me` | 当前用户 |
| `PUT` | `/api/auth/me` | 更新昵称与头像 |
| `GET/POST` | `/api/admin/users` | 管理员查看 / 创建用户 |
| `GET` | `/api/places?q=` | 地点搜索 |
| `GET/POST` | `/api/markers` | 我的足迹列表 / 创建 |
| `PUT/DELETE` | `/api/markers/{id}` | 更新 / 删除 |
| `POST` | `/api/upload` | 上传图片（需登录） |
| `POST` | `/api/shares` | 创建共享地图 |
| `GET` | `/api/shares` | 我发出的共享地图 |
| `GET` | `/api/shares/inbox` | 别人分享给我的 |
| `PUT` | `/api/shares/{id}` | 更新标题 / 公开性 / 协作权限 |
| `POST` | `/api/shares/{id}/members` | 按用户名邀请成员 |
| `DELETE` | `/api/shares/{id}/members/{userId}` | 移除成员，`me` 表示自己退出 |
| `GET` | `/api/s/{token}` | 共享地图数据（公开链接可匿名） |

足迹与上传接口需要 `Authorization: Bearer <token>`。

---

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3、Pinia、Vue Router、Element Plus、Axios、高德 JSAPI |
| 后端 | Go `net/http`、SQLite（modernc.org/sqlite）、JWT、bcrypt |
| 构建 | Vite |

---

## 截图

![Screenshot 1](./assets/readmeImages/image-20250830011233852.png)

![Screenshot 2](./assets/readmeImages/image-20250830011259609.png)

![Screenshot 3](./assets/readmeImages/image-20250830011310020.png)

![Screenshot 4](./assets/readmeImages/image-20250830011331507.png)
