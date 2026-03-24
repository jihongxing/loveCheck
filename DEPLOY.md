# LoverTrust 生产部署文档

## 一、架构总览

```
用户浏览器 ──→ :80 Nginx (frontend 容器)
                   │
                   ├── 静态文件（Vue SPA, gzip 压缩）
                   │
                   └── /api/* ──→ proxy_pass ──→ backend:8080 (Go 容器)
                                                   │
                                    ┌──────────────┼──────────────┐
                                    │              │              │
                              PostgreSQL 16    Redis 7       MinIO
                              (Hash Index)    (限流/缓存)   (证据存储)
```

全部 5 个服务通过 `docker-compose.prod.yml` 编排，一键部署。

---

## 二、项目文件结构

```
loveCheck/
├── docker-compose.yml            # 开发环境（仅基础设施）
├── docker-compose.prod.yml       # 生产环境（全部 5 个服务）
├── .env.prod.example             # 生产环境变量模板
│
├── backend/
│   ├── Dockerfile                # Go 多阶段构建
│   ├── .dockerignore
│   ├── .env                      # 开发环境变量（不入版本控制）
│   ├── .env.example              # 开发环境变量模板
│   ├── cmd/lovecheck/main.go     # 入口
│   ├── internal/
│   │   ├── bloom/bloom.go        # Bloom Filter（内存级快速排除）
│   │   ├── db/db.go              # PostgreSQL 连接 + 连接池 + 索引
│   │   ├── handler/              # API 处理器
│   │   ├── middleware/           # 限流中间件
│   │   ├── model/record.go       # 数据模型（JSONB tags）
│   │   └── storage/minio.go      # MinIO 对象存储
│   └── pkg/crypto/crypto.go      # HMAC-SHA256 确定性哈希
│
└── frontend/
    ├── Dockerfile                # Node 构建 + Nginx 运行
    ├── .dockerignore
    ├── nginx.conf                # Nginx 配置（反代 + 安全头 + gzip）
    ├── src/
    │   ├── components/           # Vue 组件
    │   ├── i18n/                 # 8 语言国际化
    │   └── App.vue
    └── package.json
```

---

## 三、前置条件

- 一台 Linux 服务器（建议 2C4G 起步）
- Docker 和 Docker Compose 已安装
- 域名已解析到服务器 IP（如需 HTTPS）

---

## 四、部署步骤

### 4.1 克隆项目

```bash
git clone <your-repo-url> /opt/lovecheck
cd /opt/lovecheck
```

### 4.2 配置环境变量

```bash
cp .env.prod.example .env
```

编辑 `.env`，**务必修改以下密钥为强随机值**：

| 变量 | 说明 | 必须修改 |
|------|------|---------|
| `PG_PASSWORD` | PostgreSQL 密码 | 是 |
| `REDIS_PASSWORD` | Redis 密码 | 是 |
| `ADMIN_SECRET` | Admin 后台登录密钥 | 是 |
| `SEARCH_PEPPER` | HMAC 哈希密钥（**部署后不可更改**） | 是 |
| `MINIO_ACCESS_KEY` | MinIO 访问密钥 | 是 |
| `MINIO_SECRET_KEY` | MinIO 秘密密钥 | 是 |
| `CORS_ORIGIN` | 前端域名，如 `https://lovertrust.com` | 生产必改 |
| `PUBLIC_PORT` | 对外端口，默认 80 | 按需 |

> **警告**：`SEARCH_PEPPER` 在有数据后**绝不能修改**，否则所有已有哈希索引将失效。

### 4.3 一键构建并启动

```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

首次构建约 2-5 分钟（下载基础镜像 + 编译 Go + npm install）。后续重启秒级完成。

### 4.4 验证服务状态

```bash
# 查看所有容器状态
docker-compose -f docker-compose.prod.yml ps

# 查看后端日志
docker-compose -f docker-compose.prod.yml logs -f backend

# 健康检查
curl http://localhost/api/v1/health
# 预期返回: {"message":"LoveCheck is running.","status":"UP"}
```

---

## 五、服务说明

### 5.1 后端 (backend)

| 项目 | 说明 |
|------|------|
| 基础镜像 | `golang:1.25-alpine` 构建 → `alpine:3.21` 运行 |
| 最终镜像大小 | ~20 MB |
| 端口 | 8080（仅容器内网） |
| 运行模式 | `GIN_MODE=release`（无 debug 日志） |
| 连接池 | 20 idle / 100 max open connections |
| Bloom Filter | 启动时预加载，支持 1000 万条记录，约 23 MB 内存 |

### 5.2 前端 (frontend)

| 项目 | 说明 |
|------|------|
| 构建工具 | Vite 8 |
| 运行容器 | Nginx Alpine |
| 端口 | 80（映射到宿主机 `PUBLIC_PORT`） |
| Gzip | 已启用（text/css/js/json/xml/svg） |
| 安全头 | X-Frame-Options / X-Content-Type-Options / X-XSS-Protection / Referrer-Policy |
| SPA 路由 | `try_files $uri $uri/ /index.html` |
| 静态资源缓存 | 30 天，`Cache-Control: public, immutable` |

### 5.3 PostgreSQL

| 项目 | 说明 |
|------|------|
| 版本 | 16 Alpine |
| 索引策略 | Hash Index（target_hash, target_local_hash）+ Partial B-Tree（仅 active 记录）+ GIN（JSONB tags） |
| 数据持久化 | Docker Volume `pgdata` |
| 健康检查 | `pg_isready` 每 5 秒 |

### 5.4 Redis

| 项目 | 说明 |
|------|------|
| 版本 | 7 Alpine |
| 用途 | 分层限流（写 10 次/分钟，读 60 次/分钟）+ 公开统计缓存（5 分钟 TTL） |
| 数据持久化 | Docker Volume `redisdata` |

### 5.5 MinIO

| 项目 | 说明 |
|------|------|
| 用途 | 举报/申诉证据图片的对象存储 |
| API 端口 | 9000（仅容器内网） |
| 控制台端口 | 9001（仅容器内网，生产建议不暴露） |
| 数据持久化 | Docker Volume `miniodata` |

---

## 六、限流策略

| 路由类型 | 限制 | 说明 |
|---------|------|------|
| 写操作 (report/appeal/vote/activate) | 10 次/分钟/IP | 防止恶意灌水 |
| 读操作 (query/check-access/platforms) | 60 次/分钟/IP | 允许正常页面交互 |
| 证据图片 (/evidence) | 无限制 | 静态资源，Referer 防护 |
| Admin (/admin/*) | 无限制 | Secret Key 认证保护 |

---

## 七、HTTPS 配置（推荐）

### 方案 A：Cloudflare（最简单）

1. 将域名 DNS 托管到 Cloudflare
2. 开启 "Proxied" 模式，自动获得 HTTPS + CDN + DDoS 防护
3. SSL 模式选择 "Flexible" 或 "Full"

### 方案 B：Nginx + Let's Encrypt

在服务器上额外安装 Certbot，修改 `nginx.conf` 添加 SSL 配置：

```bash
# 安装 Certbot
apt install certbot python3-certbot-nginx

# 申请证书
certbot --nginx -d your-domain.com
```

---

## 八、数据备份

### 自动备份脚本

```bash
#!/bin/bash
# /opt/lovecheck/backup.sh
BACKUP_DIR="/opt/backups/lovecheck"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# PostgreSQL
docker-compose -f /opt/lovecheck/docker-compose.prod.yml exec -T postgres \
  pg_dump -U lovecheck lovecheck | gzip > $BACKUP_DIR/pg_$DATE.sql.gz

# 保留最近 30 天
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

添加定时任务：

```bash
# 每天凌晨 3 点自动备份
crontab -e
0 3 * * * /opt/lovecheck/backup.sh >> /var/log/lovecheck-backup.log 2>&1
```

---

## 九、常用运维命令

```bash
# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 停止所有服务（数据保留）
docker-compose -f docker-compose.prod.yml down

# 重新构建并启动（代码更新后）
docker-compose -f docker-compose.prod.yml up -d --build

# 查看实时日志
docker-compose -f docker-compose.prod.yml logs -f backend

# 进入 PostgreSQL 命令行
docker-compose -f docker-compose.prod.yml exec postgres psql -U lovecheck

# 查看数据量
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U lovecheck -c "SELECT count(*) FROM risk_records;"

# 清理无用的 Docker 镜像
docker system prune -f
```

---

## 十、性能指标

| 数据量 | 查询延迟（含 Bloom Filter） | 说明 |
|--------|---------------------------|------|
| 10 万 | < 2 ms | Bloom 排除 90% 无效查询 |
| 100 万 | < 5 ms | Hash Index O(1) 命中 |
| 1000 万 | < 10 ms | 23 MB Bloom + Hash Index |
| 1 亿 | < 15 ms | 建议此时引入分区表 |

---

## 十一、安全检查清单

- [ ] 所有密钥已从默认值修改为强随机值
- [ ] `SEARCH_PEPPER` 已记录并安全保管（不可丢失）
- [ ] `CORS_ORIGIN` 已设置为实际域名（非 `*`）
- [ ] MinIO 控制台端口 9001 未对外暴露
- [ ] HTTPS 已配置
- [ ] 数据库备份定时任务已设置
- [ ] Admin 密钥足够复杂

---

## 十二、从 GitHub 自动部署（Deploy Key + Actions SSH）

推送 `main` 且 [CI](.github/workflows/ci.yml) 成功后，会触发 [Deploy](.github/workflows/deploy.yml)，通过 SSH 在服务器上执行 `git pull` 与 `docker compose`。需要两类 SSH 密钥，用途不同，请勿混用。

### 12.1 密钥分工

| 密钥 | 用途 | 私钥存放位置 |
|------|------|----------------|
| **仓库 Deploy Key** | 服务器从 GitHub **拉代码**（`git pull`） | 仅保存在服务器 `~/.ssh/` |
| **Actions 部署密钥** | **GitHub Actions** 登录你的服务器执行命令 | 填入仓库 Secret `DEPLOY_SSH_PRIVATE_KEY` |

### 12.2 在服务器上：配置 Deploy Key（拉代码）

在服务器上以将执行部署的用户登录（例如 `ubuntu`）：

```bash
ssh-keygen -t ed25519 -f ~/.ssh/lovecheck_repo -N "" -C "lovecheck-deploy-key"
```

将 **公钥** `~/.ssh/lovecheck_repo.pub` 全文复制到 GitHub：仓库 **Settings → Deploy keys → Add deploy key**，勾选 **Allow read access**（不要勾选 write，除非确有需要）。

配置 SSH 只对 `github.com` 使用该私钥：

```bash
cat >> ~/.ssh/config << 'EOF'
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/lovecheck_repo
  IdentitiesOnly yes
EOF
chmod 600 ~/.ssh/config ~/.ssh/lovecheck_repo
```

首次部署前克隆（路径需与下方 Secret `DEPLOY_PATH` 一致，例如 `/opt/lovecheck`）：

```bash
sudo mkdir -p /opt/lovecheck
sudo chown "$USER:$USER" /opt/lovecheck
git clone git@github.com:jihongxing/loveCheck.git /opt/lovecheck
cd /opt/lovecheck
cp .env.prod.example .env
# 编辑 .env …
docker compose -f docker-compose.prod.yml up -d --build
```

确保该用户可使用 Docker（任选其一）：`sudo usermod -aG docker "$USER"` 后重新登录，或把脚本里的 `docker compose` 改成 `sudo docker compose`。

### 12.3 在服务器上：配置 Actions 登录用的公钥

在**本机或服务器**再生成一对专给 CI 用的密钥（不要用 Deploy Key 那对）：

```bash
ssh-keygen -t ed25519 -f ./gha_lovecheck -N "" -C "github-actions-deploy"
```

将 **`gha_lovecheck.pub`** 追加到部署用户主目录：

```bash
cat gha_lovecheck.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

将 **`gha_lovecheck`（私钥）** 的完整内容（含 `BEGIN` / `END` 行）复制到 GitHub 仓库 **Settings → Secrets and variables → Actions → New repository secret**，名称：`DEPLOY_SSH_PRIVATE_KEY`。私钥文件不要上传仓库，用毕可删除本地副本。

### 12.4 GitHub 仓库 Secrets / Variables

在 **Settings → Secrets and variables → Actions** 中新增 **Secrets**：

| Name | 说明 |
|------|------|
| `DEPLOY_HOST` | 服务器公网 IP 或域名 |
| `DEPLOY_USER` | SSH 登录用户名 |
| `DEPLOY_SSH_PRIVATE_KEY` | 上一节的 CI 用私钥全文 |
| `DEPLOY_PATH` | 仓库在服务器上的绝对路径，如 `/opt/lovecheck` |
| `DEPLOY_KNOWN_HOSTS` | 在任意机器执行 `ssh-keyscan -p 22 -H <DEPLOY_HOST>` 的完整输出（多行），用于校验主机指纹 |

非 22 端口时，在 **Variables** 中增加 `DEPLOY_SSH_PORT`（例如 `2222`）；未设置则默认 `22`。

### 12.5 手动触发部署

在 GitHub 打开 **Actions → Deploy → Run workflow**，可在不推送代码时重新部署当前 `main`。

### 12.6 常见问题

- **Deploy 未运行**：确认是推送到 `main` 且 **CI** 工作流已成功结束；Pull Request 仅跑 CI，不会自动部署。
- **`git pull` 失败**：检查 Deploy Key 是否已添加、服务器 `~/.ssh/config` 是否指向正确私钥、`origin` 是否为 `git@github.com:jihongxing/loveCheck.git`（`git remote -v`）。
- **SSH 连接被拒**：检查安全组/防火墙放行 SSH 端口、`DEPLOY_KNOWN_HOSTS` 是否与当前主机密钥一致。
