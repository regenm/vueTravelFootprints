# 旅迹 · 旅行足迹地图

Vue 3 前端 + Go 后端的多用户旅行足迹服务。在地图上记录去过的地方，用照片轮播重温旅途，并通过链接与好友共享同一张地图。

本仓库基于上游 [regenm/vueTravelFootprints](https://github.com/regenm/vueTravelFootprints) 二次开发，面向家庭旅行记录场景，新增功能见下一节。

---

## 功能

> 下面第一节是上游原有功能，其后是本仓库新增的部分。

### 上游原有

- 登录后使用；未开放自行注册，账号只能由管理员创建
- 个人头像与昵称，地图和共享页都会显示记录者
- 搜索地点即可添加足迹，也可以点击地图选点
- 分类、日期、旅行笔记、多图上传
- 详情页图片轮播，点击可全屏预览
- 时间线侧栏、搜索与分类筛选
- 创建共享地图：全部或挑选部分足迹，公开链接或仅邀请
- 按用户名邀请好友查看，也可一起在同一张地图上记录
- 公开分享页无需登录即可浏览

### 本仓库新增

#### 坐标体系统一为 WGS-84

- 内置 WGS-84 / GCJ-02 双向坐标转换工具（含单元测试）
- 数据库统一存储 WGS-84 坐标，仅在地图渲染时转换为高德 GCJ-02
- 修复浏览器定位约 500 米的偏移；搜索、地图选点、浏览器定位三个入口走同一条转换链路，编辑保存后坐标不再漂移

#### 行政区划自动识别

- 创建/更新足迹时，后端自动调用高德行政区划反查，写入省份与城市的名称、区划码（markers 表新增 `province_code`、`city_code`、`province_name`、`city_name` 四列）
- 存量数据在下次编辑保存时自动补齐

#### 到访区县点亮

- 地图自动为去过的区县做半透明填色（高德 `DistrictLayer.Province` + `depth 2` + 函数式样式，按区县级行政区划码精确点亮，只渲染到访省份、仅点亮去过的区县）
- 右下角 `◉` 按钮可开关填色图层，`▦` 按钮切换卫星影像（自动叠加路网标注，点亮在卫星模式下同样生效）
- 早期按市级点亮的存量足迹会在下次编辑保存时自动补齐区县码

#### 统计扩展

- 侧栏统计新增：省份 `x/34`、城市 `x/333`、行程数

#### 行程（Trips）

- 行程的新建 / 改名 / 删除
- 足迹可归属到某次行程，详情面板与编辑表单中下拉选择
- 删除行程自动解除其下足迹的归属（足迹本身保留），分享视图行为一致

#### 视觉重制

- 主色改为蓝色 `#4877ad`，Element Plus 七档色阶同步覆盖
- 地图底图改用高德 `light` 浅色样式，与浅色 UI 协调
- 足迹分类色采用苹果系统色系（systemGreen / Brown / Orange / Indigo / Teal / Purple / Red）

#### 数据备份

- 管理员头像菜单「导出备份」一键下载 zip：内含 `data/travel.db`（`VACUUM INTO` 在线一致性快照，无需停服）与全部 `uploads/` 图片
- 恢复方式（手动）：停服 → 解压 zip，用其中的 `data/` 与 `uploads/` 覆盖运行目录 → 重新启动

#### 开发脚本

- 根目录 `./dev.sh start|stop|restart|status|log` 一键启停前后端，端口读取 `.env`，可用 `BACKEND_PORT` / `FRONTEND_PORT` 覆盖

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
| `PUT` | `/api/auth/password` | 修改自己的密码（需旧密码） |
| `GET/POST` | `/api/admin/users` | 管理员查看 / 创建用户 |
| `GET` | `/api/admin/export` | 导出备份 zip（数据库快照 + 全部图片，仅管理员） |
| `GET` | `/api/places?q=` | 地点搜索 |
| `GET/POST` | `/api/markers` | 我的足迹列表 / 创建 |
| `PUT/DELETE` | `/api/markers/{id}` | 更新 / 删除 |
| `GET/POST` | `/api/trips` | 行程列表 / 新建 |
| `PUT/DELETE` | `/api/trips/{id}` | 行程改名 / 删除（自动解除足迹归属） |
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

